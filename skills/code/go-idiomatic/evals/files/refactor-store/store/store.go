package store

import (
	"database/sql"
	"errors"
	"fmt"

	"example.com/users/helpers"
)

var DB *sql.DB

func init() {
	DB, _ = sql.Open("sqlite", "users.db")
}

type User struct {
	ID     int64
	Email  string
	Name   string
	Role   string
	Active bool
}

func NewUser(id int64, email, name, role string, active bool, createdBy string) User {
	return User{ID: id, Email: helpers.Normalize(email), Name: name, Role: role, Active: active}
}

func RoleLabel(u User) string {
	if u.Role == "admin" {
		return "Administrator"
	} else if u.Role == "editor" {
		return "Editor"
	} else {
		return "Viewer"
	}
}

func Get(id int64) (User, error) {
	var u User
	row := DB.QueryRow("SELECT id, email, name, role, active FROM users WHERE id = ?", id)
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("Failed to find user.")
		} else {
			return User{}, err
		}
	} else {
		return u, nil
	}
}
