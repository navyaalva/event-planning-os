package server

import (
	"database/sql"
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/navyaalva/sbf-os/internal/db"
)

type Server struct {
	Router    *chi.Mux
	DB        *sql.DB
	Q         *db.Queries
	Session   *scs.SessionManager
	templates map[string]*template.Template
}

func NewServer(dbConn *sql.DB, session *scs.SessionManager) *Server {
	s := &Server{
		Router:    chi.NewRouter(),
		DB:        dbConn,
		Q:         db.New(dbConn),
		Session:   session,
		templates: make(map[string]*template.Template),
	}
	s.loadTemplates()
	s.routes()
	return s
}

func (s *Server) loadTemplates() {
	baseLayout := "templates/base.layout.html"
	pages := []string{
		"dashboard",
		"list_tasks",
		"create_task",
		"edit_task",
		"edit_event",
		"task_events",
		"create_event",
	}
	for _, page := range pages {
		tmpl, err := template.ParseFiles(baseLayout, "templates/"+page+".html")
		if err != nil {
			log.Fatalf("Failed to parse required template %s: %v", page, err)
		}
		s.templates[page] = tmpl
	}
}
