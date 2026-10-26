package todo

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func NewHandler(store *Store) http.Handler {
	mux := http.NewServeMux()
	h := &handler{store: store}
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /todos", h.list)
	mux.HandleFunc("POST /todos", h.create)
	mux.HandleFunc("GET /todos/{id}", h.get)
	mux.HandleFunc("PATCH /todos/{id}", h.update)
	mux.HandleFunc("DELETE /todos/{id}", h.delete)
	return mux
}

type handler struct {
	store *Store
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "down",
			"database": "down",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"database": "up",
	})
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list todos")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"todos": todos})
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(body.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	todo, err := h.store.Create(r.Context(), body.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create todo")
		return
	}
	writeJSON(w, http.StatusCreated, todo)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	todo, err := h.store.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not get todo")
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.Title != nil && strings.TrimSpace(*body.Title) == "" {
		writeError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}
	todo, err := h.store.Update(r.Context(), id, body.Title, body.Done)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update todo")
		return
	}
	writeJSON(w, http.StatusOK, todo)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	err := h.store.Delete(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete todo")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
