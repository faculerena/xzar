package handler

import (
	"encoding/json"
	"net/http"

	"xz.ar/internal/auth"
	"xz.ar/internal/db"
	"xz.ar/internal/model"
)

type APIHandler struct {
	store  *db.Store
	domain string
	creds  *auth.Credentials
}

type shortenRequest struct {
	URL  string `json:"url"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

type shortenResponse struct {
	URL      string `json:"url"`
	Slug     string `json:"slug"`
	Type     string `json:"type"`
	ShortURL string `json:"short_url"`
}

func (h *APIHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || !h.creds.Verify(user, pass) {
		w.Header().Set("WWW-Authenticate", `Basic realm="xzar"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		jsonError(w, "url is required", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = "path"
	}
	typ := model.ShortcutType(req.Type)
	if typ != model.ShortcutPath && typ != model.ShortcutSubdomain {
		jsonError(w, "type must be 'path' or 'subdomain'", http.StatusBadRequest)
		return
	}

	slug := req.Slug
	if slug == "" {
		slug = randomSlug(8)
	}

	if err := h.store.CreateShortcut(slug, req.URL, typ); err != nil {
		jsonError(w, "slug already exists", http.StatusConflict)
		return
	}

	var shortURL string
	if typ == model.ShortcutSubdomain {
		shortURL = "https://" + slug + "." + h.domain
	} else {
		shortURL = "https://" + h.domain + "/" + slug
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shortenResponse{
		URL:      req.URL,
		Slug:     slug,
		Type:     req.Type,
		ShortURL: shortURL,
	})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
