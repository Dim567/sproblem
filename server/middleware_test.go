package server

import (
	"net/http"
	"testing"

	"github.com/Dim567/sproblem/db"
)

type DbTest struct {
	users map[int]db.User
}

func (dt *DbTest) CreateUser(user db.User) (int64, error) {
	dt.users[1] = user
	return 1, nil
}
func (dt *DbTest) GetUserById(id int64) (*db.User, error) {
	user := dt.users[int(id)]
	return &user, nil
}
func (dt *DbTest) UpdateUserById(id int64, newData map[string]any) error {
	return nil
}
func (dt *DbTest) DeleteUserById(id int64) error {
	return nil
}
func (dt *DbTest) Close() error {
	return nil
}

func TestDatabaseInjectorMiddleware(t *testing.T) {
	database := &DbTest{users: make(map[int]db.User)}
	expectedName := "John"
	user := db.User{Name: expectedName}

	innerFunc := func(database db.Database, w http.ResponseWriter, r *http.Request) {
		database.CreateUser(user)
	}
	handler := databaseInjectorMiddleware(database, innerFunc)
	handler(nil, nil)

	retrievedUser := database.users[1]

	if retrievedUser.Name != expectedName {
		t.Errorf("got %s, wanted %s", retrievedUser.Name, expectedName)
	}
}
