package rest

type ResponseEnvelope[T any] struct {
	RequestId  string `json:"request_id" binding:"required" format:"uuid" example:"669cd40f-ea15-40a8-ab03-81e704a3ecf9"`
	Status     Status `json:"status" binding:"required" example:"SUCCESS" description:"Request status"`
	StatusCode int    `json:"status_code" binding:"required" example:"200" description:"HTTP status code"`
	Details    T      `json:"details" binding:"required"`
}
