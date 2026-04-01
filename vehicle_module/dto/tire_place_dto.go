package dto

type TirePlaceCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type TirePlaceUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

type TirePlaceResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TirePlaceListResponse struct {
	Items []TirePlaceResponse `json:"items"`
	Total int64               `json:"total"`
}
