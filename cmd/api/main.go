// Command api runs the hospital middleware HTTP service.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/auth"
	"github.com/Nutto555/hospital-middleware/internal/config"
	"github.com/Nutto555/hospital-middleware/internal/repository/postgres"
	"github.com/Nutto555/hospital-middleware/internal/rest"
	"github.com/Nutto555/hospital-middleware/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		return err
	}
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	tokens := auth.NewJWT(cfg.JWTSecret, cfg.JWTTTL)
	router := rest.NewRouter(rest.Deps{
		Staff: service.NewStaff(postgres.NewHospitalRepo(pool), postgres.NewStaffRepo(pool), tokens),
	})
	return serve(ctx, ":"+cfg.Port, router)
}

// serve runs the server until ctx is cancelled, then drains connections for up to ten seconds.
func serve(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Printf("listening on %s", addr)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
