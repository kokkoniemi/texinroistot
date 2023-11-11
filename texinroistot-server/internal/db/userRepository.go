package db

import (
	"context"
	"time"
)

type userRepo struct{}

// Create implements UserRepository.
func (*userRepo) Create(user User) (*User, error) {
	panic("unimplemented")
}

// List implements UserRepository.
func (*userRepo) List(pageIndex int) ([]*User, *ListMeta, error) {
	db, err := GetDB()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancelCtx()

	var users []*User
	rows, err := db.QueryContext(ctx, "SELECT id, created_at, hash, is_admin FROM users;")
	if err != nil {
		return nil, nil, err
	}

	for rows.Next() {
		var u User
		if err = rows.Scan(&u.ID, &u.CreatedAt, &u.Hash, &u.IsAdmin); err != nil {
			return nil, nil, err
		}
		users = append(users, &u)
	}

	return users, nil, nil
}

// Remove implements UserRepository.
func (*userRepo) Remove(userHash string) error {
	panic("unimplemented")
}

// SetAdmin implements UserRepository.
func (*userRepo) SetAdmin(userHash string) (*User, error) {
	panic("unimplemented")
}

func NewUserRepository() UserRepository {
	return &userRepo{}
}
