package model

type Category struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}
