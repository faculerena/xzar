package handler

import (
	"net/http"
	"strings"

	"xz.ar/internal/db"
	"xz.ar/internal/model"
)

type RedirectHandler struct {
	store  *db.Store
	domain string
}

func (h *RedirectHandler) HandleSubdomain(w http.ResponseWriter, r *http.Request, subdomain string) {
	sc, err := h.store.GetShortcutBySlug(model.ShortcutSubdomain, subdomain)
	if err != nil || sc == nil {
		http.Redirect(w, r, "https://"+h.domain+"/", http.StatusFound)
		return
	}
	go h.store.IncrementClickCount(sc.ID)
	http.Redirect(w, r, sc.TargetURL, http.StatusFound)
}

func (h *RedirectHandler) HandlePath(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/")
	if slug == "" {
		http.Redirect(w, r, "https://"+h.domain+"/", http.StatusFound)
		return
	}
	sc, err := h.store.GetShortcutBySlug(model.ShortcutPath, slug)
	if err != nil || sc == nil {
		http.Redirect(w, r, "https://"+h.domain+"/", http.StatusFound)
		return
	}
	go h.store.IncrementClickCount(sc.ID)
	http.Redirect(w, r, sc.TargetURL, http.StatusFound)
}
