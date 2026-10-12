package user

import (
	"encoding/json"
	"fmt"
	"io"
)

type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func Decode(r io.Reader) (User, error) {
	var u User
	if err := json.NewDecoder(r).Decode(&u); err != nil {
		return User{}, fmt.Errorf("decode user: %w", err)
	}
	return u, nil
}
