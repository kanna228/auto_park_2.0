package models

import "time"

type WaybillRoutePoint struct {
	ID                  int64
	WaybillID           int64
	SeqNumber           int
	Destination         string
	ArrivalTime         *string
	HospitalizationTime *string
	LPUArrivalTime      *string
	ReleaseTime         *string
	CreatedAt           time.Time
}
