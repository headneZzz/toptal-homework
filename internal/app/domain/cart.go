package domain

import "time"

type Cart struct {
	Id        int       `db:"id"`
	UserId    int       `db:"user_id"`
	BookId    int       `db:"book_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
