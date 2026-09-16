// Command hismock stands in for the Hospital A patient API during local runs and demos.
// It answers GET /patient/search/{id} from the embedded fixtures, by national id or passport id.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/his"
)

//go:embed patients.json
var fixtures []byte

func main() {
	patients, err := load(fixtures)
	if err != nil {
		log.Fatal(err)
	}
	addr := ":" + getenv("PORT", "8081")
	srv := &http.Server{Addr: addr, Handler: handler(patients), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("hospital a mock listening on %s with %d patients", addr, len(patients))
	log.Fatal(srv.ListenAndServe())
}

func load(b []byte) ([]his.Patient, error) {
	var patients []his.Patient
	if err := json.Unmarshal(b, &patients); err != nil {
		return nil, fmt.Errorf("fixtures: %w", err)
	}
	return patients, nil
}

func handler(patients []his.Patient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /patient/search/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		for _, p := range patients {
			if matches(p, id) {
				writeJSON(w, http.StatusOK, p)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "patient not found"})
	})
	return mux
}

func matches(p his.Patient, id string) bool {
	return (p.NationalID != nil && *p.NationalID == id) || (p.PassportID != nil && *p.PassportID == id)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
