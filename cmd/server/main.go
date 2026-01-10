package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/navyaalva/sbf-os/internal/server"
)

func main() {
	_ = godotenv.Load() // Ignore error for prod compatibility

	dbConn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	// --- NEW: Session Manager Setup ---
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Store = postgresstore.New(dbConn)
	sessionManager.Cookie.Secure = os.Getenv("ENV") == "production" // Secure only in prod
	sessionManager.Cookie.HttpOnly = true
	// ----------------------------------

	// Pass sessionManager to the server
	srv := server.NewServer(dbConn, sessionManager)

	log.Println("🚀 SBF-OS running on :8080")
	if err := http.ListenAndServe(":8080", sessionManager.LoadAndSave(srv.Router)); err != nil {
		log.Fatal(err)
	}
}
