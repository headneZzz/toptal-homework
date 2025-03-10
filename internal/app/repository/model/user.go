package model

import "time"

type User struct {
	Id           int       `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Admin        bool      `db:"admin"`
	CreatedAt    time.Time `db:"created_at"`
	// TODO rename to cart_updated_at
	UpdatedAt time.Time `db:"updated_at"`
}
