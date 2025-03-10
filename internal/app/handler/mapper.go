package handler

import (
	"log"
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
)

func toBookWithId(request model.BookUpdateRequest) domain.Book {
	b, err := domain.NewBook(request.Id, request.Title, request.Year, request.Author, request.Price, request.Stock, request.CategoryId)
	if err != nil {
		log.Fatalf("failed to convert BookUpdateRequest to domain.Book: %v", err)
	}
	return b
}

func toBook(request model.BookCreateRequest) domain.Book {
	b, err := domain.NewBook(1, request.Title, request.Year, request.Author, request.Price, request.Stock, request.CategoryId)
	if err != nil {
		log.Fatalf("failed to convert BookCreateRequest to domain.Book: %v", err)
	}
	return b
}

func toBookResponse(book domain.Book) model.BookResponse {
	return model.BookResponse{
		Title:      book.Title(),
		Year:       book.Year(),
		Author:     book.Author(),
		Price:      book.Price(),
		Stock:      book.Stock(),
		CategoryId: book.CategoryId(),
	}
}

func toBooksResponse(books []domain.Book) []model.BookResponse {
	responses := make([]model.BookResponse, len(books))
	for i, book := range books {
		responses[i] = toBookResponse(book)
	}
	return responses
}

func toCategoryResponse(category domain.Category) model.CategoryResponse {
	return model.CategoryResponse{
		Id:   category.Id(),
		Name: category.Name(),
	}
}

func toCategoriesResponse(categories []domain.Category) []model.CategoryResponse {
	responses := make([]model.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = toCategoryResponse(category)
	}
	return responses
}

func toCategory(request model.CategoryRequest) (domain.Category, error) {
	var category domain.Category
	err := category.SetName(request.Name)
	return category, err
}
