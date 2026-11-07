package router_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aidar-khabibullin/go-service-template/pkg/router"
)

var errDown = errors.New("down")

func get(h http.Handler, path string) int {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))

	return rec.Code
}

func TestProbes(t *testing.T) {
	t.Parallel()

	ok := router.Check{Name: "ok", Fn: func(context.Context) error { return nil }}
	bad := router.Check{Name: "bad", Fn: func(context.Context) error { return errDown }}

	if code := get(router.New(ok), "/ready"); code != http.StatusOK {
		t.Errorf("ready with healthy deps: got %d", code)
	}

	if code := get(router.New(ok, bad), "/ready"); code != http.StatusServiceUnavailable {
		t.Errorf("ready with a dead dep: got %d", code)
	}

	// liveness must not depend on dependencies
	if code := get(router.New(bad), "/live"); code != http.StatusOK {
		t.Errorf("live: got %d", code)
	}
}
