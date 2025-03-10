package repository

import (
	"log"
	"log/slog"
	"toptal/internal/app/domain"
	"toptal/internal/app/repository/model"
)

func toDomainBook(book model.Book) domain.Book {
	b, err := domain.NewBook(book.Id, book.Title, book.Year, book.Author, book.Price, book.Stock, book.CategoryId)
	if err != nil {
		log.Fatalf("failed to map model.Book to domain.Book: %v", err)
	}
	return b
}

func toDomainBooks(books []model.Book) []domain.Book {
	domainBooks := make([]domain.Book, len(books))
	for i, book := range books {
		domainBooks[i] = toDomainBook(book)
	}
	return domainBooks
}

func toDomainCategory(category model.Category) (domain.Category, error) {
	return domain.NewCategory(category.Id, category.Name)
}

func toDomainCategories(categories []model.Category) ([]domain.Category, error) {
	domains := make([]domain.Category, len(categories))
	var err error
	for i, cat := range categories {
		domains[i], err = toDomainCategory(cat)
		if err != nil {
			slog.Error("failed to map model.Category to domain.Category", "error", err)
			return nil, err
		}
	}
	return domains, nil
}

func toDomainUser(user model.User) (domain.User, error) {
	return domain.NewUser(user.Id, user.Username, user.PasswordHash, user.Admin)
}

func toModelUser(user domain.User) model.User {
	return model.User{
		Id:           user.Id(),
		Username:     user.Username(),
		PasswordHash: user.PasswordHash(),
		Admin:        user.Admin(),
	}
}
