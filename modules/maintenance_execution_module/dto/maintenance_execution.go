package dto

import "time"

type MaintenanceExecutionRequest struct {
	ScheduleID          int64  `json:"schedule_id" example:"5"`
	Board               string `json:"board" example:"30"`
	Action              string `json:"action" example:"serviced_replaced"`
	Comment             string `json:"comment" example:"replaced spark plugs"`
	ResponsibleMechanic string `json:"responsible_mechanic" example:"Azamat Rakhimov"`
}

type MaintenanceExecutionBulkItemRequest struct {
	Board   string `json:"board" example:"30"`
	Action  string `json:"action" example:"serviced"`
	Comment string `json:"comment" example:""`
}

type MaintenanceExecutionBulkRequest struct {
	ScheduleID          int64                                 `json:"schedule_id" example:"5"`
	ResponsibleMechanic string                                `json:"responsible_mechanic" example:"Azamat Rakhimov"`
	Items               []MaintenanceExecutionBulkItemRequest `json:"items"`
}

type MaintenanceExecutionResponse struct {
	ID                  int64     `json:"id" example:"1"`
	ScheduleID          int64     `json:"schedule_id" example:"5"`
	Board               string    `json:"board" example:"30"`
	Action              string    `json:"action" example:"serviced_replaced"`
	Comment             *string   `json:"comment" example:"replaced spark plugs"`
	ResponsibleMechanic *string   `json:"responsible_mechanic" example:"Azamat Rakhimov"`
	CreatedAt           time.Time `json:"created_at" example:"2026-03-20T13:10:00Z"`
	UpdatedAt           time.Time `json:"updated_at" example:"2026-03-20T13:10:00Z"`
}

type MaintenanceExecutionListResponse struct {
	Items  []MaintenanceExecutionResponse `json:"items"`
	Total  int64                          `json:"total" example:"1"`
	Limit  int                            `json:"limit" example:"50"`
	Offset int                            `json:"offset" example:"0"`
}
