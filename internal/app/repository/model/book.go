package model

type Book struct {
	Id         int    `db:"id"`
	Title      string `db:"title"`
	Year       int    `db:"year"`
	Author     string `db:"author"`
	Price      int    `db:"price"`
	Stock      int    `db:"stock"`
	CategoryId int    `db:"category_id"`
}
