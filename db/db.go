package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type Database interface {
	CreateUser(user User) (int64, error)
	GetUserById(id int64) (*User, error)
	UpdateUserById(id int64, newData map[string]any) error
	DeleteUserById(id int64) error
	Close() error
}

type dbConnection struct {
	driver *sql.DB
}

// TODO: config should come from file/cli options/vault/...
// It may have other than string format
func CreateConnection(config string) (Database, error) {
	configStr := "user=postgres dbname=test password=password sslmode=disable"
	driver, err := sql.Open("postgres", configStr)
	if err != nil {
		return nil, err
	}
	err = driver.Ping()
	if err != nil {
		return nil, err
	}

	return &dbConnection{driver}, nil
}

func (db *dbConnection) CreateUser(user User) (int64, error) {
	var userId int64
	err := db.driver.QueryRow(
		"INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id",
		user.Name, user.Email, user.Age,
	).Scan(&userId)
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func (db *dbConnection) GetUserById(id int64) (*User, error) {
	user := User{}
	err := db.driver.QueryRow(
		"SELECT id, name, email, age FROM users WHERE id = $1 LIMIT 1",
		id,
	).Scan(&user.Id, &user.Name, &user.Email, &user.Age)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return &user, err
	}

	return &user, nil
}

func (db *dbConnection) UpdateUserById(id int64, newData map[string]any) error {
	fields := ""
	values := []any{id}
	i := 2
	for key, val := range newData {
		fields += fmt.Sprintf("%s=$%d,", key, i)
		values = append(values, val)
		i++
	}
	fields = strings.TrimRight(fields, ",")
	queryTemplate := fmt.Sprintf("UPDATE users SET %s WHERE id = $1", fields)
	_, err := db.driver.Exec(queryTemplate, values...)
	return err
}

func (db *dbConnection) DeleteUserById(id int64) error {
	_, err := db.driver.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (db *dbConnection) Close() error {
	return db.driver.Close()
}
