package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"xz.ar/internal/auth"
	"xz.ar/internal/config"
	"xz.ar/internal/db"
	"xz.ar/internal/handler"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed templates/*.html templates/**/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	if len(os.Args) >= 4 && os.Args[1] == "--genpass" {
		genpass(os.Args[2], os.Args[3])
		return
	}

	cfg := config.Load()

	if cfg.SessionSecret == "" {
		log.Fatal("XZAR_SESSION_SECRET is required")
	}

	creds, err := auth.LoadCredentials(cfg.CredentialsFile)
	if err != nil {
		log.Fatalf("load credentials: %v", err)
	}

	migrationSQL, err := migrationsFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatalf("read migration: %v", err)
	}
	db.SetMigrations(string(migrationSQL))

	store, err := db.New(filepath.Join(cfg.DataDir, "xzar.db"))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer store.Close()

	os.MkdirAll(filepath.Join(cfg.DataDir, "uploads"), 0755)

	sm := auth.NewSessionManager(cfg.SessionSecret, creds)
	templates := parseTemplates()
	staticSub, _ := fs.Sub(staticFS, "static")
	h := handler.NewRouter(cfg, store, sm, templates, staticSub)

	srv := &http.Server{Addr: cfg.Addr, Handler: h}

	go func() {
		log.Printf("listening on %s (domain: %s)", cfg.Addr, cfg.Domain)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("shutdown complete")
}

func genpass(username, password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}
	data, _ := json.MarshalIndent(map[string]string{
		"username": username,
		"password": string(hash),
	}, "", "  ")
	fmt.Println(string(data))
}

func parseTemplates() map[string]*template.Template {
	layout := template.Must(template.ParseFS(templateFS, "templates/layout.html"))

	parse := func(files ...string) *template.Template {
		t := template.Must(layout.Clone())
		return template.Must(t.ParseFS(templateFS, files...))
	}

	return map[string]*template.Template{
		"login":         parse("templates/login.html"),
		"dashboard":     parse("templates/admin/dashboard.html"),
		"shortcut_form": parse("templates/admin/shortcut_form.html"),
		"images":        parse("templates/admin/images.html"),
		"carousel":      parse("templates/public/carousel.html"),
	}
}
