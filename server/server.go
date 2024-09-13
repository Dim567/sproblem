package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Dim567/sproblem/db"
	"github.com/gorilla/mux"
)

var server *http.Server

// TODO: config should come from file/cli options/vault/...
// It may have other than string format
func Start(config string) error {
	if server != nil {
		return fmt.Errorf("server already started")
	}

	dbConn, err := db.CreateConnection("mock config")
	if err != nil {
		return fmt.Errorf("database error: %+v", err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/users", loggerMiddleware("createUser", databaseInjectorMiddleware(dbConn, createUser))).Methods("POST")
	r.HandleFunc("/users/{id}", loggerMiddleware("getUser", databaseInjectorMiddleware(dbConn, getUser))).Methods("GET")
	r.HandleFunc("/users/{id}", loggerMiddleware("updateUser", databaseInjectorMiddleware(dbConn, updateUser))).Methods("PUT")
	r.HandleFunc("/users/{id}", loggerMiddleware("deleteUser", databaseInjectorMiddleware(dbConn, deleteUser))).Methods("DELETE")
	http.Handle("/", r)

	server = &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	return server.ListenAndServe()
}

func createUser(dbConn db.Database, w http.ResponseWriter, r *http.Request) {
	userData, err := io.ReadAll(r.Body)
	if err != nil {
		msg := "failed to read user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}
	user := db.User{}
	err = json.Unmarshal(userData, &user)
	if err != nil {
		msg := "failed to parse user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}
	if user.Age <= 0 || user.Name == "" || user.Email == "" {
		msg := "wrong user data"
		logError(msg, nil)
		sendError(w, msg, http.StatusBadRequest)
		return
	}

	userId, err := dbConn.CreateUser(user)
	if err != nil {
		msg := "failed to create user"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, userId)
}

func getUser(dbConn db.Database, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr, ok := vars["id"]
	if !ok {
		msg := "user ID is not provided"
		logError(msg, nil)
		sendError(w, msg, http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		msg := "wrong user ID"
		logError(msg, err)
		sendError(w, msg, http.StatusBadRequest)
		return
	}

	user, err := dbConn.GetUserById(int64(userId))
	if user == nil {
		msg := "user not found"
		logError(msg, nil)
		sendError(w, msg, http.StatusNotFound)
		return
	}
	if err != nil {
		msg := "failed to get user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(user)
	if err != nil {
		msg := "unable to serialize user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(data))
}

func updateUser(dbConn db.Database, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr, ok := vars["id"]
	if !ok {
		msg := "user ID is not provided"
		logError(msg, nil)
		sendError(w, msg, http.StatusBadRequest)
		return
	}

	userData, err := io.ReadAll(r.Body)
	if err != nil {
		msg := "failed to read user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}

	var newData map[string]any
	err = json.Unmarshal([]byte(userData), &newData)
	if err != nil {
		msg := "failed to parse user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		msg := "wrong user ID"
		logError(msg, err)
		sendError(w, msg, http.StatusBadRequest)
		return
	}

	user, err := dbConn.GetUserById(int64(userId))
	if user == nil {
		msg := "user not found"
		logError(msg, nil)
		sendError(w, msg, http.StatusNotFound)
		return
	}
	if err != nil {
		msg := "failed to get user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}

	err = dbConn.UpdateUserById(int64(userId), newData)
	if err != nil {
		msg := "failed to update user"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, userId)
}

func deleteUser(dbConn db.Database, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIdStr, ok := vars["id"]
	if !ok {
		msg := "user ID is not provided"
		logError(msg, nil)
		sendError(w, msg, http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		msg := "wrong user ID"
		logError(msg, err)
		sendError(w, msg, http.StatusBadRequest)
		return
	}

	user, err := dbConn.GetUserById(int64(userId))
	if user == nil {
		msg := "user not found"
		logError(msg, nil)
		sendError(w, msg, http.StatusNotFound)
		return
	}
	if err != nil {
		msg := "failed to get user data"
		logError(msg, err)
		sendError(w, msg, http.StatusInternalServerError)
		return
	}

	err = dbConn.DeleteUserById(int64(userId))
	if err != nil {
		msg := "failed to delete user"
		log.Println(fmt.Errorf(msg), err)
		fmt.Fprint(w, msg)
		return
	}
	fmt.Fprint(w, userId)
}

func logError(msg string, err error) {
	log.Println(fmt.Errorf(msg), err)
}

func sendError(w http.ResponseWriter, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprint(w, msg)
}
