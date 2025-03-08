package handler

import (
	"encoding/json"
	"net/http"
	"toptal/internal/app/auth"
	"toptal/internal/app/handler/model"
)

// @Summary Get user's cart
// @Description Get the current user's shopping cart contents
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {array} model.BookResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart [get]
func (s *Server) handleGetCart(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cart, err := s.cartService.GetCart(r.Context(), userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responses []model.BookResponse
	for _, book := range cart {
		res := toBookResponse(book)
		responses = append(responses, res)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Add book to cart
// @Description Add a book to the current user's shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Param request body model.AddToCartRequest true "Book to add to cart"
// @Success 202 {string} string "Accepted"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Book not found"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart/add [post]
func (s *Server) handleAddToCart(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var cartRequest model.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&cartRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.cartService.AddToCart(r.Context(), userId, cartRequest.BookId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Book not found in cart"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart/remove [post]
func (s *Server) handleRemoveFromCart(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var cartRequest model.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&cartRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.cartService.RemoveFromCart(r.Context(), userId, cartRequest.BookId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 422 {string} string "Insufficient stock"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /cart/purchase [post]
func (s *Server) handlePurchase(w http.ResponseWriter, r *http.Request) {
	userId, err := auth.GetUserId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.cartService.Purchase(r.Context(), userId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
