package domain

import "fmt"

type Book struct {
	id         int
	title      string
	year       int
	author     string
	price      int
	stock      int
	categoryId int
}

func NewBook(id int, title string, year int, author string, price int, stock int, categoryId int) (Book, error) {
	book := Book{}
	if err := book.SetID(id); err != nil {
		return book, err
	}
	if err := book.SetTitle(title); err != nil {
		return book, err
	}
	if err := book.SetYear(year); err != nil {
		return book, err
	}
	if err := book.SetAuthor(author); err != nil {
		return book, err
	}
	if err := book.SetPrice(price); err != nil {
		return book, err
	}
	if err := book.SetStock(stock); err != nil {
		return book, err
	}
	if err := book.SetCategoryId(categoryId); err != nil {
		return book, err
	}
	return book, nil
}

// Getter methods

func (b *Book) Id() int {
	return b.id
}

func (b *Book) Title() string {
	return b.title
}

func (b *Book) Year() int {
	return b.year
}

func (b *Book) Author() string {
	return b.author
}

func (b *Book) Price() int {
	return b.price
}

func (b *Book) Stock() int {
	return b.stock
}

func (b *Book) CategoryId() int {
	return b.categoryId
}

// Setter methods with validations

func (b *Book) SetID(id int) error {
	if id <= 0 {
		return fmt.Errorf("id must be a positive integer")
	}
	b.id = id
	return nil
}

func (b *Book) SetTitle(title string) error {
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	b.title = title
	return nil
}

func (b *Book) SetYear(year int) error {
	if year <= 0 {
		return fmt.Errorf("year must be greater than zero")
	}
	b.year = year
	return nil
}

func (b *Book) SetAuthor(author string) error {
	if author == "" {
		return fmt.Errorf("author cannot be empty")
	}
	b.author = author
	return nil
}

func (b *Book) SetPrice(price int) error {
	if price < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	b.price = price
	return nil
}

func (b *Book) SetStock(stock int) error {
	if stock < 0 {
		return fmt.Errorf("stock cannot be negative")
	}
	b.stock = stock
	return nil
}

func (b *Book) SetCategoryId(categoryId int) error {
	if categoryId <= 0 {
		return fmt.Errorf("categoryId must be a positive integer")
	}
	b.categoryId = categoryId
	return nil
}
