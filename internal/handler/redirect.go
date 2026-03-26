package handler

import (
	"io"
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
	h.sendRedirect(w, r, sc.TargetURL)
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
	h.sendRedirect(w, r, sc.TargetURL)
}

func (h *RedirectHandler) sendRedirect(w http.ResponseWriter, r *http.Request, target string) {
	if strings.HasPrefix(r.UserAgent(), "curl/") {
		resp, err := http.Get(target)
		if err != nil {
			http.Error(w, "failed to fetch target", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for _, key := range []string{"Content-Type", "Content-Length"} {
			if v := resp.Header.Get(key); v != "" {
				w.Header().Set(key, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}
