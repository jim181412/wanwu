package web

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"institute.supwisdom.com/authx-demo-cas-go/internal/cas"
	"institute.supwisdom.com/authx-demo-cas-go/session"
)

type Handler struct {
	contextPath    string
	authMode       string
	sessions       *session.Manager
	indexTemplate  *template.Template
	hasDemoArchive bool
}

func NewHandler(contextPath string, authMode string, sessions *session.Manager) (*Handler, error) {
	indexTemplate, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return nil, err
	}

	_, err = os.Stat("demo/authx-demo-cas-go.zip")

	return &Handler{
		contextPath:    contextPath,
		authMode:       authMode,
		sessions:       sessions,
		indexTemplate:  indexTemplate,
		hasDemoArchive: err == nil,
	}, nil
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	locals := map[string]interface{}{
		"contextPath":    h.contextPath,
		"authMode":       h.authMode,
		"casUser":        cas.NewCasUser(),
		"hasDemoArchive": h.hasDemoArchive,
	}

	sess := h.sessions.SessionStart(w, r)
	if casUser, ok := sess.Get("casUser").(*cas.CasUser); ok {
		locals["casUser"] = casUser
	}

	if err := h.indexTemplate.Execute(w, locals); err != nil {
		log.Printf("render index failed: %v", err)
		http.Error(w, "render index failed", http.StatusInternalServerError)
	}
}
