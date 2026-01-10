package server

import (
	"database/sql"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/navyaalva/sbf-os/internal/db"
)

type Server struct {
	Router  *chi.Mux
	DB      *sql.DB
	Q       *db.Queries
	Session *scs.SessionManager // <--- Added
}

func NewServer(dbConn *sql.DB, session *scs.SessionManager) *Server {
	s := &Server{
		Router:  chi.NewRouter(),
		DB:      dbConn,
		Q:       db.New(dbConn),
		Session: session,
	}
	s.routes()
	return s
}
