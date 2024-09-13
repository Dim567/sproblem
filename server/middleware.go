package server

import (
	"log"
	"net/http"
	"time"

	"github.com/Dim567/sproblem/db"
)

func databaseInjectorMiddleware(
	database db.Database,
	f func(dbs db.Database, w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(database, w, r)
	}
}

// Logs the time each request takes from start to finish
func loggerMiddleware(
	handlerName string,
	f func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		f(w, r)
		end := time.Now()
		log.Printf("Duration of the %s request = %s", handlerName, end.Sub(start))
	}
}
