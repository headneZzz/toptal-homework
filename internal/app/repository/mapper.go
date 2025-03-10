package repository

import (
	"toptal/internal/app/domain"
	"toptal/internal/app/repository/model"
)

func toDomainBook(book model.Book) domain.Book {
	return domain.Book{
		Id:         book.Id,
		Title:      book.Title,
		Year:       book.Year,
		Author:     book.Author,
		Price:      book.Price,
		Stock:      book.Stock,
		CategoryId: book.CategoryId,
	}
}

func toDomainBooks(books []model.Book) []domain.Book {
	domainBooks := make([]domain.Book, len(books))
	for i, book := range books {
		domainBooks[i] = toDomainBook(book)
	}
	return domainBooks
}

func toDomainCategory(category model.Category) domain.Category {
	return domain.Category{
		Id:   category.Id,
		Name: category.Name,
	}
}

func toDomainCategories(categories []model.Category) []domain.Category {
	domainCategories := make([]domain.Category, len(categories))
	for i, category := range categories {
		domainCategories[i] = toDomainCategory(category)
	}
	return domainCategories
}

func toDomainUser(user model.User) domain.User {
	return domain.User{
		Id:           user.Id,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Admin:        user.Admin,
	}
}

func toModelUser(user domain.User) model.User {
	return model.User{
		Id:           user.Id,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Admin:        user.Admin,
	}
}
