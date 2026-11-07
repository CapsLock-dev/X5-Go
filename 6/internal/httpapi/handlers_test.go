package httpapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"homework6/internal/httpapi"
	"homework6/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// До реализации health этот тест должен быть красным.
func TestHealth(t *testing.T) {
	h := httpapi.NewRouter(model.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status in body=%q", body.Status)
	}
}

func TestCreateOrder(t *testing.T) {
	correctHeader := http.Header{"Content-Type": {"application/json"}}
	tests := []struct {
		name   string
		header http.Header
		body   string
		status int
		want   model.Order
	}{
		{"correct customer", correctHeader, `{"customer":"Test"}`, 201, model.Order{ID: 1, Customer: "Test", Status: "new"}},
		{"incorrect customer", correctHeader, `{"customer":" "}`, 400, model.Order{}},
		{"second customer", correctHeader, `{"customer":"Test2"}`, 201, model.Order{ID: 2, Customer: "Test2", Status: "new"}},
		{"broken json", correctHeader, `{"customer":"Test2"`, 400, model.Order{}},
	}

	h := httpapi.NewRouter(model.NewStore())

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(tt.body))
		req.Header = tt.header

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tt.status {
			t.Fatalf("%s) status=%d, want %d", tt.name, rec.Code, tt.status)
		}
		if rec.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("%s) Content-Type=%q", tt.name, rec.Header().Get("Content-Type"))
		}
		var body model.Order
		decoder := json.NewDecoder(rec.Body)
		if err := decoder.Decode(&body); err != nil {
			t.Fatalf("%s) %s", tt.name, err.Error())
		}
		if tt.status == 201 && body != tt.want {
			t.Fatalf("%s) body=%v want=%v", tt.name, body, tt.want)
		}
		if tt.status == 201 && rec.Header().Get("Location") != fmt.Sprintf("/orders/%d", tt.want.ID) {
			t.Fatalf("%s) location=%q want=%s", tt.name, rec.Header().Get("Location"), fmt.Sprintf("/orders/%d", tt.want.ID))
		}
	}
}

func TestGetOrder(t *testing.T) {
	correctHeader := http.Header{"Content-Type": {"application/json"}}
	tests := []struct {
		name   string
		header http.Header
		id     any
		status int
		want   model.Order
	}{
		{"order exists", correctHeader, 1, 200, model.Order{ID: 1, Customer: "customer1", Status: "new"}},
		{"another order exists", correctHeader, 3, 200, model.Order{ID: 3, Customer: "customer2", Status: "new"}},
		{"order doesnt exist", correctHeader, 100, 404, model.Order{}},
		{"invalid negative id", correctHeader, -1, 400, model.Order{}},
		{"invalid NaN id", correctHeader, "asd", 400, model.Order{}},
	}

	store := model.NewStore()
	store.Create("customer1")
	store.Create("customer1")
	store.Create("customer2")
	h := httpapi.NewRouter(store)

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/orders/%v", tt.id), strings.NewReader(""))
		req.Header = tt.header

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tt.status {
			t.Fatalf("%s) status=%d, want %d", tt.name, rec.Code, tt.status)
		}
		if rec.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("%s) Content-Type=%q", tt.name, rec.Header().Get("Content-Type"))
		}
		if tt.status == 200 {
			var body model.Order
			decoder := json.NewDecoder(rec.Body)
			if err := decoder.Decode(&body); err != nil {
				t.Fatalf("%s) %s", tt.name, err.Error())
			}
			if body != tt.want {
				t.Fatalf("%s) body=%v want=%v", tt.name, body, tt.want)
			}
		}
	}
}

func TestRecovery(t *testing.T) {
	r := httpapi.NewRouter(model.NewStore())
	r.(*chi.Mux).Get("/panic-test", func(w http.ResponseWriter, r *http.Request) {
		panic("panic")
	})
	req := httptest.NewRequest(http.MethodGet, "/panic-test", strings.NewReader(""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status=%d, want %d", rec.Code, 500)
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "Internal server error" {
		t.Fatalf("error in body=%q", body.Error)
	}
}
