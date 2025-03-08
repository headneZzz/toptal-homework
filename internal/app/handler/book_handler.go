package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"toptal/internal/app/handler/model"
	"toptal/internal/pkg/validator"
)

// BookHandler handles HTTP requests for books
type BookHandler struct {
	bookService BookService
}

// NewBookHandler creates a new BookHandler
func NewBookHandler(bookService BookService) *BookHandler {
	return &BookHandler{bookService}
}

// @Summary Get book by ID
// @Description Get a book's details by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} model.BookResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /book/{id} [get]
func (s *Server) handleGetBookById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	book, err := s.bookService.GetBookById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := toBookResponse(book)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Get available books
// @Description Get a list of all available books, optionally filtered by category IDs
// @Tags books
// @Accept json
// @Produce json
// @Param categoryId query []int false "Category IDs to filter by"
// @Success 200 {array} model.BookResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /book [get]
func (s *Server) handleGetBooks(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query()["categoryId"]
	ids := make([]int, len(idsStr))
	for i, v := range idsStr {
		id, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ids[i] = id
	}

	books, err := s.bookService.GetAvailableBooks(r.Context(), ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]model.BookResponse, len(books))
	for i, book := range books {
		response[i] = toBookResponse(book)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Create a new book
// @Description Create a new book with the provided details
// @Tags books
// @Accept json
// @Produce json
// @Param book body model.BookRequest true "Book details"
// @Success 201 {object} model.BookResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /book [post]
func (s *Server) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var bookRequest model.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&bookRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	if err := validator.Validate(bookRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	book := toBook(bookRequest, 0)
	if err := s.bookService.CreateBook(r.Context(), book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := toBookResponse(book)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Update a book
// @Description Update an existing book's details
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param book body model.BookRequest true "Updated book details"
// @Success 200 {object} model.BookResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /book/{id} [put]
func (s *Server) handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		WriteValidationError(w, err)
		return
	}

	var bookRequest model.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&bookRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	if err := validator.Validate(bookRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	book := toBook(bookRequest, id)
	if err := s.bookService.UpdateBook(r.Context(), book); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := toBookResponse(book)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Delete a book
// @Description Delete a book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /book/{id} [delete]
func (s *Server) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.bookService.DeleteBook(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
