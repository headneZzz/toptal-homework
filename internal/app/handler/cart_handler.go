package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"toptal/internal/app/auth"
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
)

// @Summary Get user's cart
// @Description Get the current user's shopping cart contents
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {array} model.BookResponse
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart [get]
func (s *Server) handleGetCart(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid User ID", err.Error(), r.URL.Path)
		return
	}

	books, err := s.cartService.GetCart(r.Context(), userId)
	if err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	response := toBooksResponse(books)
	writeResponseOK(w, response)
}

// @Summary Add book to cart
// @Description Add a book to the current user's shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Param request body model.AddToCartRequest true "Book to add to cart"
// @Success 202 {string} string "Accepted"
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 404 {object} model.ProblemDetail "Book not found"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart/add [post]
func (s *Server) handleAddToCart(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid User ID", err.Error(), r.URL.Path)
		return
	}

	var cartRequest model.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&cartRequest); err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid Request", err.Error(), r.URL.Path)
		return
	}

	if err := s.cartService.AddToCart(r.Context(), userId, cartRequest.BookId); err != nil {
		if errors.Is(err, domain.ErrBookNotFound) {
			model.NotFound(w, "Book not found", r.URL.Path)
		} else if errors.Is(err, domain.ErrBookOutOfStock) {
			model.ValidationError(w, "Book out of stock", r.URL.Path)
		} else {
			model.InternalServerError(w, r.URL.Path)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// @Summary Remove book from cart
// @Description Remove a book from the current user's shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Param request body model.AddToCartRequest true "Book to remove from cart"
// @Success 202 {string} string "Accepted"
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 404 {object} model.ProblemDetail "Book not found in cart"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart/remove [post]
func (s *Server) handleRemoveFromCart(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid User ID", err.Error(), r.URL.Path)
		return
	}

	var cartRequest model.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&cartRequest); err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid Request", err.Error(), r.URL.Path)
		return
	}

	if err := s.cartService.RemoveFromCart(r.Context(), userId, cartRequest.BookId); err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// @Summary Purchase cart
// @Description Purchase all books in the current user's shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Success 202 {string} string "Accepted"
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 422 {object} model.ProblemDetail "Insufficient stock"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart/purchase [post]
func (s *Server) handlePurchase(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid User ID", err.Error(), r.URL.Path)
		return
	}

	if err := s.cartService.Purchase(r.Context(), userId); err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
