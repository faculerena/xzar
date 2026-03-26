package handler

import (
	"fmt"
	"html/template"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"xz.ar/internal/db"
	"xz.ar/internal/model"

	"github.com/google/uuid"
)

type AdminHandler struct {
	store     *db.Store
	domain    string
	dataDir   string
	templates map[string]*template.Template
}

func (h *AdminHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.templates["login"].Execute(w, map[string]any{
		"Error": r.URL.Query().Get("error") != "",
	})
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	shortcuts, err := h.store.ListShortcuts()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cfg, err := h.store.GetHomepageConfig()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.templates["dashboard"].Execute(w, map[string]any{
		"Shortcuts": shortcuts,
		"Homepage":  cfg,
		"Domain":    h.domain,
		"Flash":     r.URL.Query().Get("flash"),
	})
}

func (h *AdminHandler) ShortcutForm(w http.ResponseWriter, r *http.Request) {
	h.templates["shortcut_form"].Execute(w, map[string]any{
		"Domain": h.domain,
	})
}

func (h *AdminHandler) ShortcutEditForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	sc, err := h.store.GetShortcutByID(id)
	if err != nil || sc == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	h.templates["shortcut_form"].Execute(w, map[string]any{
		"Shortcut": sc,
		"Domain":   h.domain,
	})
}

func randomSlug(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

var reservedPathSlugs = map[string]bool{
	"admin": true, "static": true, "uploads": true,
	"favicon.ico": true, "robots.txt": true,
}

var reservedSubdomainSlugs = map[string]bool{
	"www": true,
}

func (h *AdminHandler) CreateShortcut(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.FormValue("slug"))
	targetURL := strings.TrimSpace(r.FormValue("target_url"))
	typ := model.ShortcutType(r.FormValue("type"))

	if targetURL == "" || (typ != model.ShortcutPath && typ != model.ShortcutSubdomain) {
		http.Redirect(w, r, "/admin/shortcuts/new?error=invalid", http.StatusSeeOther)
		return
	}

	if slug == "" {
		slug = randomSlug(8)
	}

	if typ == model.ShortcutPath && reservedPathSlugs[slug] {
		http.Redirect(w, r, "/admin/shortcuts/new?error=reserved", http.StatusSeeOther)
		return
	}
	if typ == model.ShortcutSubdomain && reservedSubdomainSlugs[slug] {
		http.Redirect(w, r, "/admin/shortcuts/new?error=reserved", http.StatusSeeOther)
		return
	}

	if err := h.store.CreateShortcut(slug, targetURL, typ); err != nil {
		http.Redirect(w, r, "/admin/shortcuts/new?error=exists", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?flash=created", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateShortcut(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	slug := strings.TrimSpace(r.FormValue("slug"))
	targetURL := strings.TrimSpace(r.FormValue("target_url"))
	typ := model.ShortcutType(r.FormValue("type"))

	if slug == "" || targetURL == "" || (typ != model.ShortcutPath && typ != model.ShortcutSubdomain) {
		http.Redirect(w, r, fmt.Sprintf("/admin/shortcuts/%d/edit?error=invalid", id), http.StatusSeeOther)
		return
	}

	if err := h.store.UpdateShortcut(id, slug, targetURL, typ); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/shortcuts/%d/edit?error=failed", id), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin?flash=updated", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteShortcut(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	h.store.DeleteShortcut(id)
	http.Redirect(w, r, "/admin?flash=deleted", http.StatusSeeOther)
}

func (h *AdminHandler) UpdateHomepage(w http.ResponseWriter, r *http.Request) {
	mode := model.HomepageMode(r.FormValue("mode"))
	redirectURL := strings.TrimSpace(r.FormValue("redirect_url"))

	if mode != model.HomepageModeRedirect && mode != model.HomepageModeCarousel {
		http.Redirect(w, r, "/admin?flash=error", http.StatusSeeOther)
		return
	}
	h.store.UpdateHomepageConfig(mode, redirectURL)
	http.Redirect(w, r, "/admin?flash=homepage_updated", http.StatusSeeOther)
}

func (h *AdminHandler) ImagesPage(w http.ResponseWriter, r *http.Request) {
	images, err := h.store.ListCarouselImages()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.templates["images"].Execute(w, map[string]any{
		"Images": images,
		"Flash":  r.URL.Query().Get("flash"),
	})
}

func (h *AdminHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Redirect(w, r, "/admin/images?flash=error", http.StatusSeeOther)
		return
	}
	defer file.Close()

	mime := header.Header.Get("Content-Type")
	allowed := map[string]string{
		"image/gif":  ".gif",
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/webp": ".webp",
	}
	ext, ok := allowed[mime]
	if !ok {
		http.Redirect(w, r, "/admin/images?flash=invalid_type", http.StatusSeeOther)
		return
	}

	filename := uuid.New().String() + ext
	dst, err := os.Create(filepath.Join(h.dataDir, "uploads", filename))
	if err != nil {
		http.Redirect(w, r, "/admin/images?flash=error", http.StatusSeeOther)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	h.store.CreateCarouselImage(filename, header.Filename, mime)
	http.Redirect(w, r, "/admin/images?flash=uploaded", http.StatusSeeOther)
}

func (h *AdminHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	img, err := h.store.DeleteCarouselImage(id)
	if err == nil && img != nil {
		os.Remove(filepath.Join(h.dataDir, "uploads", img.Filename))
	}
	http.Redirect(w, r, "/admin/images?flash=deleted", http.StatusSeeOther)
}

func (h *AdminHandler) ReorderImages(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	idsStr := r.Form["ids"]
	var ids []int64
	for _, s := range idsStr {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	h.store.ReorderCarouselImages(ids)
	http.Redirect(w, r, "/admin/images?flash=reordered", http.StatusSeeOther)
}
