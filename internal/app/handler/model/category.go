package model

type CategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type CategoryResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
