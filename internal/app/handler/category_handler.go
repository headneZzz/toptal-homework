package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
	"toptal/internal/pkg/validator"
)

// @Summary Get category by ID
// @Description Get a category's details by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} model.CategoryResponse
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 404 {object} model.ProblemDetail "Not Found"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Router /category/{id} [get]
func (s *Server) handleGetCategoryById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid Category ID", err.Error(), r.URL.Path)
		return
	}

	category, err := s.categoryService.GetCategoryById(r.Context(), id)
	if err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	response := toCategoryResponse(category)
	writeResponseOK(w, response)
}

// @Summary Get all categories
// @Description Get a list of all available categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} model.CategoryResponse
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Router /category [get]
func (s *Server) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.categoryService.GetCategories(r.Context())
	if err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	response := toCategoriesResponse(categories)
	writeResponseOK(w, response)
}

// @Summary Create a new category
// @Description Create a new category with the provided details
// @Tags categories
// @Accept json
// @Produce json
// @Param category body model.Category true "Category details"
// @Success 201
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /category [post]
func (s *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var categoryRequest model.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		model.InvalidRequest(w, err.Error(), r.URL.Path)
		return
	}

	if err := validator.Validate(categoryRequest); err != nil {
		model.ValidationError(w, err.Error(), r.URL.Path)
		return
	}

	category := toCategory(categoryRequest)
	if err := s.categoryService.CreateCategory(r.Context(), category); err != nil {
		switch {
		case errors.Is(err, domain.ErrAlreadyExists):
			model.AlreadyExists(w, "Category Already Exists", r.URL.Path)
		case errors.Is(err, domain.ErrForbidden):
			model.Forbidden(w, err.Error(), r.URL.Path)
		default:
			model.InternalServerError(w, r.URL.Path)
		}
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
// @Success 200
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 404 {object} model.ProblemDetail "Not Found"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /category [put]
func (s *Server) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	var categoryRequest model.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid Request", err.Error(), r.URL.Path)
		return
	}

	if err := validator.Validate(categoryRequest); err != nil {
		model.ValidationError(w, err.Error(), r.URL.Path)
		return
	}

	category := toCategory(categoryRequest)
	if err := s.categoryService.UpdateCategory(r.Context(), category); err != nil {
		model.InternalServerError(w, r.URL.Path)
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
// @Success 200
// @Failure 400 {object} model.ProblemDetail "Bad Request"
// @Failure 401 {object} model.ProblemDetail "Unauthorized"
// @Failure 404 {object} model.ProblemDetail "Not Found"
// @Failure 500 {object} model.ProblemDetail "Internal Server Error"
// @Security ApiKeyAuth
// @Router /category/{id} [delete]
func (s *Server) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		model.WriteProblemDetail(w, http.StatusBadRequest, "Invalid Category ID", err.Error(), r.URL.Path)
		return
	}

	if err := s.categoryService.DeleteCategory(r.Context(), id); err != nil {
		model.InternalServerError(w, r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusOK)
}
