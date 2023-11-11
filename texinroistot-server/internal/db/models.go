package db

import "time"

type User struct {
	ID        int
	CreatedAt time.Time `json:"createdAt"`
	Hash      string    `json:"hash"`
	IsAdmin   bool      `json:"isAdmin"`
}
