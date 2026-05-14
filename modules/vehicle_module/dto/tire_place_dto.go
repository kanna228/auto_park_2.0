package dto

type TirePlaceCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type TirePlaceUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

type TirePlaceListQuery struct {
	Name   string
	Limit  int
	Offset int
	SortBy string
	Order  string
}

type TirePlaceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TirePlaceListResponse struct {
	Items  []TirePlaceResponse `json:"items"`
	Total  int64               `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}
