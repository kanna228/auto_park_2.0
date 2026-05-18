package dto

type WaybillRoutePointCreateRequest struct {
	SeqNumber           int     `json:"seq_number" binding:"required"`
	Destination         string  `json:"destination" binding:"required"`
	ArrivalTime         *string `json:"arrival_time,omitempty"`
	HospitalizationTime *string `json:"hospitalization_time,omitempty"`
	LPUArrivalTime      *string `json:"lpu_arrival_time,omitempty"`
	ReleaseTime         *string `json:"release_time,omitempty"`
}

type WaybillRoutePointUpdateRequest struct {
	SeqNumber           *int    `json:"seq_number,omitempty"`
	Destination         *string `json:"destination,omitempty"`
	ArrivalTime         *string `json:"arrival_time,omitempty"`
	HospitalizationTime *string `json:"hospitalization_time,omitempty"`
	LPUArrivalTime      *string `json:"lpu_arrival_time,omitempty"`
	ReleaseTime         *string `json:"release_time,omitempty"`
}

type WaybillRoutePointResponse struct {
	ID                  int64   `json:"id"`
	WaybillID           int64   `json:"waybill_id"`
	SeqNumber           int     `json:"seq_number"`
	Destination         string  `json:"destination"`
	ArrivalTime         *string `json:"arrival_time,omitempty"`
	HospitalizationTime *string `json:"hospitalization_time,omitempty"`
	LPUArrivalTime      *string `json:"lpu_arrival_time,omitempty"`
	ReleaseTime         *string `json:"release_time,omitempty"`
	CreatedAt           string  `json:"created_at"`
}
