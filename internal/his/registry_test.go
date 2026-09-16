package his

import (
	"context"
	"testing"
)

type nopClient struct{}

func (nopClient) SearchPatient(context.Context, string) (Patient, error) {
	return Patient{}, ErrNotFound
}

func TestRegistryFor(t *testing.T) {
	r := Registry{"hospital-a": nopClient{}}
	if _, ok := r.For("hospital-a"); !ok {
		t.Fatal("hospital-a must be found")
	}
	if _, ok := r.For("hospital-b"); ok {
		t.Fatal("hospital-b has no client")
	}
	if _, ok := Registry(nil).For("hospital-a"); ok {
		t.Fatal("nil registry has no clients")
	}
}
