package handler

import (
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
)

func toBook(request model.BookRequest, id int) domain.Book {
	return domain.Book{
		Id:         id,
		Title:      request.Title,
		Year:       request.Year,
		Author:     request.Author,
		Price:      request.Price,
		Stock:      request.Stock,
		CategoryId: request.CategoryId,
	}
}

func toBookResponse(book domain.Book) model.BookResponse {
	return model.BookResponse{
		Title:      book.Title,
		Year:       book.Year,
		Author:     book.Author,
		Price:      book.Price,
		Stock:      book.Stock,
		CategoryId: book.CategoryId,
	}
}

func toCategoryRequest(category domain.Category) model.Category {
	return model.Category{Name: category.Name}
}

func toCategory(category model.Category) domain.Category {
	return domain.Category{Name: category.Name}
}
