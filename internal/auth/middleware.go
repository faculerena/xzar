package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	cookieName = "xzar_session"
	maxAge     = 24 * time.Hour
)

type SessionManager struct {
	secret []byte
	creds  *Credentials
}

func NewSessionManager(secret string, creds *Credentials) *SessionManager {
	return &SessionManager{
		secret: []byte(secret),
		creds:  creds,
	}
}

func (sm *SessionManager) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if !sm.creds.Verify(username, password) {
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}

	ts := fmt.Sprintf("%d", time.Now().Unix())
	sig := sm.sign(username + "|" + ts)
	value := username + "|" + ts + "|" + sig

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/admin",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (sm *SessionManager) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/admin",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (sm *SessionManager) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !sm.IsAuthenticated(r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (sm *SessionManager) IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return false
	}
	parts := strings.SplitN(cookie.Value, "|", 3)
	if len(parts) != 3 {
		return false
	}
	username, tsStr, sig := parts[0], parts[1], parts[2]

	if sig != sm.sign(username+"|"+tsStr) {
		return false
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return false
	}
	if time.Since(time.Unix(ts, 0)) > maxAge {
		return false
	}
	return true
}

func (sm *SessionManager) sign(data string) string {
	mac := hmac.New(sha256.New, sm.secret)
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
