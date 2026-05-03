package httpapi

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/CyberGeo335/pz6-web-security/internal/auth"
	"github.com/CyberGeo335/pz6-web-security/internal/store"
)

type Handler struct {
	store        *store.Store
	cookieSecure bool

	profileTmpl  *template.Template
	helloTmpl    *template.Template
	commentsTmpl *template.Template
}

func NewHandler(s *store.Store, cookieSecure bool) (*Handler, error) {
	profileTmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		return nil, err
	}

	helloTmpl, err := template.ParseFiles("templates/hello.html")
	if err != nil {
		return nil, err
	}

	commentsTmpl, err := template.ParseFiles("templates/comments.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		store:        s,
		cookieSecure: cookieSecure,
		profileTmpl:  profileTmpl,
		helloTmpl:    helloTmpl,
		commentsTmpl: commentsTmpl,
	}, nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID, err := auth.RandomToken(16)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	csrfToken, err := auth.RandomToken(16)
	if err != nil {
		http.Error(w, "failed to create csrf token", http.StatusInternalServerError)
		return
	}

	h.store.Save(&store.UserProfile{
		SessionID: sessionID,
		Name:      "Студент",
		CSRFToken: csrfToken,
	})

	auth.SetSessionCookie(w, sessionID, h.cookieSecure)
	http.Redirect(w, r, "/profile", http.StatusFound)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID, err := auth.ReadSessionCookie(r)
	if err == nil {
		h.store.Delete(sessionID)
	}

	auth.ClearSessionCookie(w, h.cookieSecure)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<p>Сессия завершена.</p><p><a href="/login">Войти снова</a></p>`))
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	profile, ok := h.currentProfile(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := struct {
			Name      string
			CSRFToken string
		}{
			Name:      profile.Name,
			CSRFToken: profile.CSRFToken,
		}

		if err := h.profileTmpl.Execute(w, data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		if !validCSRF(r.FormValue("csrf_token"), profile.CSRFToken) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		sessionID := profile.SessionID
		h.store.UpdateName(sessionID, name)
		if err := h.rotateCSRFToken(sessionID); err != nil {
			http.Error(w, "failed to rotate csrf token", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/hello", http.StatusFound)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	profile, ok := h.currentProfile(w, r)
	if !ok {
		return
	}

	data := struct {
		Name string
	}{
		Name: profile.Name,
	}

	if err := h.helloTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Comments(w http.ResponseWriter, r *http.Request) {
	profile, ok := h.currentProfile(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.renderComments(w, profile)

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		if !validCSRF(r.FormValue("csrf_token"), profile.CSRFToken) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		text := strings.TrimSpace(r.FormValue("text"))
		if text == "" {
			http.Error(w, "comment is required", http.StatusBadRequest)
			return
		}

		h.store.AddComment(store.Comment{
			Author:    profile.Name,
			Text:      text,
			CreatedAt: time.Now(),
		})

		if err := h.rotateCSRFToken(profile.SessionID); err != nil {
			http.Error(w, "failed to rotate csrf token", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/comments", http.StatusFound)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) currentProfile(w http.ResponseWriter, r *http.Request) (*store.UserProfile, bool) {
	sessionID, err := auth.ReadSessionCookie(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}

	profile, ok := h.store.Get(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusUnauthorized)
		return nil, false
	}

	return profile, true
}

func (h *Handler) rotateCSRFToken(sessionID string) error {
	csrfToken, err := auth.RandomToken(16)
	if err != nil {
		return err
	}
	h.store.UpdateCSRFToken(sessionID, csrfToken)
	return nil
}

func (h *Handler) renderComments(w http.ResponseWriter, profile *store.UserProfile) {
	data := struct {
		Name      string
		CSRFToken string
		Comments  []store.Comment
	}{
		Name:      profile.Name,
		CSRFToken: profile.CSRFToken,
		Comments:  h.store.Comments(),
	}

	if err := h.commentsTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

func validCSRF(tokenFromForm, tokenFromSession string) bool {
	return tokenFromForm != "" && tokenFromForm == tokenFromSession
}
