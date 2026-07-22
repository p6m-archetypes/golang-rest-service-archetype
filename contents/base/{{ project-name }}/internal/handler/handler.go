// Package handler serves the service's REST API: full CRUD over {{ PrefixName }} at
// /api/v1/{{ prefix-name }}s, the p6m platform's standard surface.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"{{ module_path }}/internal/repository"
)

// {{ PrefixName }} is the wire shape of the entity.
type {{ PrefixName }} struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// {{ PrefixName }}Request is the create/update payload.
type {{ PrefixName }}Request struct {
	DisplayName string `json:"displayName"`
}

type handler struct {
	store *repository.Store
}

// New builds the service router.
func New(store *repository.Store) http.Handler {
	h := &handler{store: store}

	r := chi.NewRouter()
	r.Use(requestLogger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1/{{ prefix-name }}s", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Get("/{id}", h.get)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})

	return r
}

// requestLogger logs requests through slog so log output honors LOGGING_STRUCTURED.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var req {{ PrefixName }}Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	e, err := h.store.Create(r.Context(), req.DisplayName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toWire(e))
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := []{{ PrefixName }}{}
	for _, e := range items {
		out = append(out, toWire(e))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	e, err := h.store.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toWire(e))
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	var req {{ PrefixName }}Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	e, err := h.store.Update(r.Context(), chi.URLParam(r, "id"), req.DisplayName)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toWire(e))
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toWire(e repository.{{ PrefixName }}) {{ PrefixName }} {
	return {{ PrefixName }}{ID: e.ID, DisplayName: e.DisplayName}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}
