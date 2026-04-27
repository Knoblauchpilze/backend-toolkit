package rest

type ResponseEnvelope[T any] struct {
	RequestId string `json:"requestId"`
	Status    string `json:"status"`
	Details   T      `json:"details"`
}
