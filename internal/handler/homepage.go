package handler

import (
	"html/template"
	"net/http"

	"xz.ar/internal/db"
	"xz.ar/internal/model"
)

type HomepageHandler struct {
	store     *db.Store
	domain    string
	templates map[string]*template.Template
}

func (h *HomepageHandler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.store.GetHomepageConfig()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	switch cfg.Mode {
	case model.HomepageModeRedirect:
		if cfg.RedirectURL == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("<!DOCTYPE html><html><head><title>xz.ar</title></head><body><h1>xz.ar</h1></body></html>"))
			return
		}
		redirectOrProxy(w, r, cfg.RedirectURL)

	case model.HomepageModeCarousel:
		images, err := h.store.ListCarouselImages()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		h.templates["carousel"].Execute(w, map[string]any{
			"Images": images,
			"Domain": h.domain,
		})
	}
}
