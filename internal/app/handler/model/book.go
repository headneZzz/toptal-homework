package model

type BookRequest struct {
	Title      string `json:"title" validate:"required,min=1,max=255"`
	Year       int    `json:"year" validate:"required,min=1800,max=2100"`
	Author     string `json:"author" validate:"required,min=1,max=255"`
	Price      int    `json:"price" validate:"required,min=0"`
	Stock      int    `json:"stock" validate:"required,min=0"`
	CategoryId int    `json:"category_id" validate:"required,min=1"`
}

type BookResponse struct {
	Title      string `json:"title"`
	Year       int    `json:"year"`
	Author     string `json:"author"`
	Price      int    `json:"price"`
	Stock      int    `json:"stock"`
	CategoryId int    `json:"category_id"`
}
