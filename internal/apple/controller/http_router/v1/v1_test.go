package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	v1 "github.com/aidar-khabibullin/go-service-template/internal/apple/controller/http_router/v1"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/entity"
	"github.com/aidar-khabibullin/go-service-template/internal/apple/usecase/usecasetest"
	"github.com/google/uuid"
)

func newServer(env *usecasetest.Env) http.Handler {
	h := v1.New(env.UC)

	r := chi.NewRouter()
	r.Post("/create_apple", h.CreateApple)
	r.Post("/create_pineapple", h.CreatePineApple)
	r.Get("/get_apple/{id}", h.GetApple)
	r.Put("/update_apple/{id}", h.UpdateApple)
	r.Delete("/delete_apple/{id}", h.DeleteApple)

	return r
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	return rec
}

func TestCreateApple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want int
	}{
		{"ok", `{"name":"fuji"}`, http.StatusOK},
		{"empty name", `{"name":""}`, http.StatusBadRequest},
		{"broken json", `{`, http.StatusBadRequest},
		{"unknown field", `{"name":"a","evil":1}`, http.StatusBadRequest},
		{"too large", `{"name":"` + strings.Repeat("a", 2<<20) + `"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := do(newServer(usecasetest.NewEnv()), http.MethodPost, "/create_apple", tt.body)
			if rec.Code != tt.want {
				t.Fatalf("want %d, got %d: %s", tt.want, rec.Code, rec.Body)
			}

			if tt.want == http.StatusOK && !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
				t.Errorf("bad content type %q", rec.Header().Get("Content-Type"))
			}
		})
	}
}

func TestCreateApple_InternalError(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()
	env.Postgres.Err["CreateApple"] = usecasetest.ErrInjected

	rec := do(newServer(env), http.MethodPost, "/create_apple", `{"name":"a"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}

func TestGetApple(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()
	apple := entity.Apple{ID: uuid.New(), Name: "gala", Status: entity.StatusNew}
	env.Postgres.Apples[apple.ID] = apple

	srv := newServer(env)

	rec := do(srv, http.MethodGet, "/get_apple/"+apple.ID.String(), "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"name":"gala"`) {
		t.Fatalf("got %d: %s", rec.Code, rec.Body)
	}

	if rec = do(srv, http.MethodGet, "/get_apple/"+uuid.NewString(), ""); rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}

	if rec = do(srv, http.MethodGet, "/get_apple/not-a-uuid", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestUpdateApple(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()
	apple := entity.Apple{ID: uuid.New(), Name: "old", Status: entity.StatusNew}
	env.Postgres.Apples[apple.ID] = apple

	srv := newServer(env)
	path := "/update_apple/" + apple.ID.String()

	if rec := do(srv, http.MethodPut, path, `{"name":"new","status":"success"}`); rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", rec.Code, rec.Body)
	}

	if got := env.Postgres.Apples[apple.ID]; got.Name != "new" || got.Status != entity.StatusSuccess {
		t.Errorf("not updated: %+v", got)
	}

	if rec := do(srv, http.MethodPut, path, `{"name":"new","status":"bogus"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}

	if rec := do(srv, http.MethodPut, "/update_apple/"+uuid.NewString(), `{"name":"n","status":"new"}`); rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestDeleteApple(t *testing.T) {
	t.Parallel()

	env := usecasetest.NewEnv()
	apple := entity.Apple{ID: uuid.New(), Name: "x"}
	env.Postgres.Apples[apple.ID] = apple

	srv := newServer(env)
	path := "/delete_apple/" + apple.ID.String()

	if rec := do(srv, http.MethodDelete, path, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}

	if rec := do(srv, http.MethodDelete, path, ""); rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestCreatePineApple(t *testing.T) {
	t.Parallel()

	srv := newServer(usecasetest.NewEnv())

	if rec := do(srv, http.MethodPost, "/create_pineapple", `{"name":"p"}`); rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body)
	}

	if rec := do(srv, http.MethodPost, "/create_pineapple", `{"name":""}`); rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}
