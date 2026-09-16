// Package hospitala talks to Hospital A's patient API.
package hospitala

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Nutto555/hospital-middleware/internal/his"
)

// Code is the hospital code the client is registered under.
const Code = "hospital-a"

// Client calls GET {baseURL}/patient/search/{id}.
type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: httpClient}
}

// SearchPatient looks up a patient by national id or passport id. An unknown id is his.ErrNotFound.
func (c *Client) SearchPatient(ctx context.Context, id string) (his.Patient, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/patient/search/"+url.PathEscape(id), nil)
	if err != nil {
		return his.Patient{}, fmt.Errorf("hospital a: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return his.Patient{}, fmt.Errorf("hospital a: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var p his.Patient
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			return his.Patient{}, fmt.Errorf("hospital a: decode response: %w", err)
		}
		if p.PatientHN == "" {
			return his.Patient{}, errors.New("hospital a: response has no patient_hn")
		}
		return p, nil
	case http.StatusNotFound:
		return his.Patient{}, his.ErrNotFound
	default:
		return his.Patient{}, fmt.Errorf("hospital a: unexpected status %d", resp.StatusCode)
	}
}
