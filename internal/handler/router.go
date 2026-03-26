package handler

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"xz.ar/internal/auth"
	"xz.ar/internal/config"
	"xz.ar/internal/db"
)

func NewRouter(cfg *config.Config, store *db.Store, sm *auth.SessionManager, templates map[string]*template.Template, staticFS fs.FS) http.Handler {
	redirect := &RedirectHandler{store: store, domain: cfg.Domain}
	homepage := &HomepageHandler{store: store, domain: cfg.Domain, templates: templates}
	admin := &AdminHandler{store: store, domain: cfg.Domain, dataDir: cfg.DataDir, templates: templates}

	// Admin mux (handles everything under /admin)
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /admin/login", admin.LoginPage)
	adminMux.HandleFunc("POST /admin/login", sm.Login)
	adminMux.HandleFunc("POST /admin/logout", sm.Logout)

	// Protected admin routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /admin", admin.Dashboard)
	protectedMux.HandleFunc("GET /admin/shortcuts/new", admin.ShortcutForm)
	protectedMux.HandleFunc("POST /admin/shortcuts", admin.CreateShortcut)
	protectedMux.HandleFunc("GET /admin/shortcuts/{id}/edit", admin.ShortcutEditForm)
	protectedMux.HandleFunc("POST /admin/shortcuts/{id}", admin.UpdateShortcut)
	protectedMux.HandleFunc("POST /admin/shortcuts/{id}/delete", admin.DeleteShortcut)
	protectedMux.HandleFunc("POST /admin/homepage", admin.UpdateHomepage)
	protectedMux.HandleFunc("GET /admin/images", admin.ImagesPage)
	protectedMux.HandleFunc("POST /admin/images", admin.UploadImage)
	protectedMux.HandleFunc("POST /admin/images/{id}/delete", admin.DeleteImage)
	protectedMux.HandleFunc("POST /admin/images/reorder", admin.ReorderImages)

	protected := sm.RequireAuth(protectedMux)

	staticHandler := http.StripPrefix("/static/", http.FileServerFS(staticFS))
	uploadsHandler := http.StripPrefix("/uploads/", http.FileServer(http.Dir(filepath.Join(cfg.DataDir, "uploads"))))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := stripPort(r.Host)
		subdomain := extractSubdomain(host, cfg.Domain)

		if subdomain != "" && subdomain != "www" {
			redirect.HandleSubdomain(w, r, subdomain)
			return
		}

		path := r.URL.Path

		// Static files
		if strings.HasPrefix(path, "/static/") {
			staticHandler.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(path, "/uploads/") {
			uploadsHandler.ServeHTTP(w, r)
			return
		}

		// Admin routes
		if strings.HasPrefix(path, "/admin") {
			if path == "/admin/login" || path == "/admin/logout" {
				adminMux.ServeHTTP(w, r)
			} else {
				protected.ServeHTTP(w, r)
			}
			return
		}

		// Homepage
		if path == "/" {
			homepage.HandleRoot(w, r)
			return
		}

		// Path-based shortcut
		redirect.HandlePath(w, r)
	})
}

func stripPort(host string) string {
	if i := strings.LastIndex(host, ":"); i != -1 {
		return host[:i]
	}
	return host
}

func extractSubdomain(host, domain string) string {
	suffix := "." + domain
	if strings.HasSuffix(host, suffix) {
		return strings.TrimSuffix(host, suffix)
	}
	return ""
}
