package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
	"toptal/internal/pkg/validator"
)

// @Summary Get book by ID
// @Description Get a book's details by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} model.BookResponse
// @Failure 400 {object} errors.ProblemDetail "Bad Request"
// @Failure 404 {object} errors.ProblemDetail "Not Found"
// @Failure 500 {object} errors.ProblemDetail "Internal Server Error"
// @Router /book/{id} [get]
func (s *Server) handleGetBookById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid ID", err.Error(), r.URL.Path)
		return
	}

	book, err := s.bookService.GetBookById(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			model.NotFound(w, "Book Not Found", r.URL.Path)
		} else {
			slog.Error("error getting book by id", "error", err)
			model.InternalServerError(w, r.URL.Path)
		}
		return
	}

	response := toBookResponse(book)
	writeResponseOK(w, response)
}

// @Summary Get available books
// @Description Get a list of all available books, optionally filtered by category IDs
// @Tags books
// @Accept json
// @Produce json
// @Param categoryId query []int false "Category IDs to filter by"
// @Success 200 {array} model.BookResponse
// @Failure 400 {object} errors.ProblemDetail "Bad Request"
// @Failure 500 {object} errors.ProblemDetail "Internal Server Error"
// @Router /book [get]
func (s *Server) handleGetBooks(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query()["categoryId"]
	ids := make([]int, len(idsStr))
	for i, v := range idsStr {
		id, err := strconv.Atoi(v)
		if err != nil {
			model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid Category ID", err.Error(), r.URL.Path)
			return
		}
		ids[i] = id
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	books, err := s.bookService.GetAvailableBooks(r.Context(), ids, limit, offset)
	if err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	response := toBooksResponse(books)
	writeResponseOK(w, response)
}

// @Summary Create a new book
// @Description Create a new book with the provided details
// @Tags books
// @Accept json
// @Produce json
// @Param book body model.BookCreateRequest true "Book details"
// @Success 201 {object} model.BookResponse
// @Failure 400 {object} errors.ProblemDetail "Bad Request"
// @Failure 401 {object} errors.ProblemDetail "Unauthorized"
// @Failure 500 {object} errors.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /book [post]
func (s *Server) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var bookRequest model.BookCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&bookRequest); err != nil {
		model.InvalidRequest(w, err.Error(), r.URL.Path)
		return
	}

	if err := validator.Validate(bookRequest); err != nil {
		model.ValidationError(w, err.Error(), r.URL.Path)
		return
	}

	book := toBook(bookRequest)
	if err := s.bookService.CreateBook(r.Context(), book); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			model.AlreadyExists(w, "Book Already Exists", r.URL.Path)
		} else if errors.Is(err, domain.ErrInvalidCategory) {
			model.InvalidRequest(w, "Invalid Category ID", r.URL.Path)
		} else {
			slog.Error("error creating book", "error", err)
			model.InternalServerError(w, r.URL.Path)
		}
		return
	}

	response := toBookResponse(book)
	writeResponseCreated(w, response)
}

// @Summary Update a book
// @Description Update an existing book's details
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param book body model.BookUpdateRequest true "Updated book details"
// @Success 200 {object} model.BookResponse
// @Failure 400 {object} errors.ProblemDetail "Bad Request"
// @Failure 401 {object} errors.ProblemDetail "Unauthorized"
// @Failure 404 {object} errors.ProblemDetail "Not Found"
// @Failure 500 {object} errors.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /book/{id} [put]
func (s *Server) handleUpdateBook(w http.ResponseWriter, r *http.Request) {
	var bookRequest model.BookUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&bookRequest); err != nil {
		model.InvalidRequest(w, err.Error(), r.URL.Path)
		return
	}

	if err := validator.Validate(bookRequest); err != nil {
		model.ValidationError(w, err.Error(), r.URL.Path)
		return
	}

	book := toBookWithId(bookRequest)
	if err := s.bookService.UpdateBook(r.Context(), book); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			model.NotFound(w, "Book Not Found", r.URL.Path)
		} else {
			slog.Error("error updating book", "error", err)
			model.InternalServerError(w, r.URL.Path)
		}
		return
	}

	response := toBookResponse(book)
	writeResponseOK(w, response)
}

// @Summary Delete a book
// @Description Delete a book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {string} string "OK"
// @Failure 400 {object} errors.ProblemDetail "Bad Request"
// @Failure 401 {object} errors.ProblemDetail "Unauthorized"
// @Failure 404 {object} errors.ProblemDetail "Not Found"
// @Failure 500 {object} errors.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /book/{id} [delete]
func (s *Server) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		model.InvalidRequest(w, "Invalid Book ID", r.URL.Path)
		return
	}

	if err := s.bookService.DeleteBook(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			model.NotFound(w, "Book Not Found", r.URL.Path)
		} else {
			slog.Error("error deleting book", "error", err)
			model.InternalServerError(w, r.URL.Path)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
