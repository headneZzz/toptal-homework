package handler

import (
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
)

func toBookWithId(request model.BookRequest, id int) domain.Book {
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

func toBook(request model.BookRequest) domain.Book {
	return domain.Book{
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

func toBooksResponse(books []domain.Book) []model.BookResponse {
	bookResponses := make([]model.BookResponse, len(books))
	for i, book := range books {
		bookResponses[i] = toBookResponse(book)
	}
	return bookResponses
}

func toCategoryResponse(category domain.Category) model.CategoryResponse {
	return model.CategoryResponse{
		Id:   category.Id,
		Name: category.Name,
	}
}

func toCategoriesResponse(categories []domain.Category) []model.CategoryResponse {
	categoryResponses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = toCategoryResponse(category)
	}
	return categoryResponses
}

func toCategory(category model.CategoryRequest) domain.Category {
	return domain.Category{Name: category.Name}
}
