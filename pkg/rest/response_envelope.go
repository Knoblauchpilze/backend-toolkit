package rest

type ResponseEnvelope[T any] struct {
	RequestId string `json:"requestId" format:"uuid" example:"669cd40f-ea15-40a8-ab03-81e704a3ecf9"`
	Status    Status `json:"status" enum:"SUCCESS,ERROR" example:"SUCCESS" description:"Request status"`
	Details   T      `json:"details"`
}
