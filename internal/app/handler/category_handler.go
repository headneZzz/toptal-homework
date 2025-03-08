package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"toptal/internal/app/handler/model"
	"toptal/internal/pkg/validator"
)

// @Summary Get category by ID
// @Description Get a category's details by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} model.Category
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /category/{id} [get]
func (s *Server) handleGetCategoryById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := s.categoryService.GetCategoryById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := toCategoryRequest(category)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Get all categories
// @Description Get a list of all available categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} model.Category
// @Failure 500 {string} string "Internal Server Error"
// @Router /category [get]
func (s *Server) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.categoryService.GetCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]model.Category, len(categories))
	for i, category := range categories {
		response[i] = toCategoryRequest(category)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Create a new category
// @Description Create a new category with the provided details
// @Tags categories
// @Accept json
// @Produce json
// @Param category body model.Category true "Category details"
// @Success 201 {object} model.Category
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /category [post]
func (s *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var categoryRequest model.Category
	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	if err := validator.Validate(categoryRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	category := toCategory(categoryRequest)
	if err := s.categoryService.CreateCategory(r.Context(), category); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// @Summary Update a category
// @Description Update an existing category's details
// @Tags categories
// @Accept json
// @Produce json
// @Param category body model.Category true "Updated category details"
// @Success 200 {object} model.Category
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /category [put]
func (s *Server) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	var categoryRequest model.Category
	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	if err := validator.Validate(categoryRequest); err != nil {
		WriteValidationError(w, err)
		return
	}

	category := toCategory(categoryRequest)
	if err := s.categoryService.UpdateCategory(r.Context(), category); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// @Summary Delete a category
// @Description Delete a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Not Found"
// @Failure 500 {string} string "Internal Server Error"
// @Security ApiKeyAuth
// @Router /category/{id} [delete]
func (s *Server) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.categoryService.DeleteCategory(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
