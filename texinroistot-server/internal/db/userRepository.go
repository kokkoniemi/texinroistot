package db

import (
	"database/sql"
	"fmt"
)

type userRepo struct{}

// Create implements UserRepository.
func (*userRepo) Create(user User) (*User, error) {
	rows, err := Query(
		`INSERT INTO users (hash, is_admin)
		 VALUES ($1, $2)
		 ON CONFLICT (hash) DO UPDATE
		 SET is_admin = users.is_admin OR EXCLUDED.is_admin
		 RETURNING id, created_at, hash, is_admin;`,
		user.Hash,
		user.IsAdmin,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("failed to create or update user")
	}

	var created User
	if err = rows.Scan(&created.ID, &created.CreatedAt, &created.Hash, &created.IsAdmin); err != nil {
		return nil, err
	}

	return &created, nil
}

// List implements UserRepository.
func (*userRepo) List(pageIndex int) ([]*User, *ListMeta, error) {
	rows, err := Query("SELECT id, created_at, hash, is_admin FROM users ORDER BY created_at ASC;")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		var u User
		if err = rows.Scan(&u.ID, &u.CreatedAt, &u.Hash, &u.IsAdmin); err != nil {
			return nil, nil, err
		}
		users = append(users, &u)
	}

	return users, nil, nil
}

// ReadByHash implements UserRepository.
func (*userRepo) ReadByHash(userHash string) (*User, error) {
	rows, err := Query(
		"SELECT id, created_at, hash, is_admin FROM users WHERE hash = $1 LIMIT 1;",
		userHash,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var user User
	if err = rows.Scan(&user.ID, &user.CreatedAt, &user.Hash, &user.IsAdmin); err != nil {
		return nil, err
	}

	return &user, nil
}

// Remove implements UserRepository.
func (*userRepo) Remove(userHash string) error {
	_, err := Execute("DELETE FROM users WHERE hash = $1;", userHash)
	return err
}

// SetAdmin implements UserRepository.
func (*userRepo) SetAdmin(userHash string) (*User, error) {
	rows, err := Query(
		`UPDATE users
		 SET is_admin = true
		 WHERE hash = $1
		 RETURNING id, created_at, hash, is_admin;`,
		userHash,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	var updated User
	if err = rows.Scan(&updated.ID, &updated.CreatedAt, &updated.Hash, &updated.IsAdmin); err != nil {
		return nil, err
	}

	return &updated, nil
}

func NewUserRepository() UserRepository {
	return &userRepo{}
}
