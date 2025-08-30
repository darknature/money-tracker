package response

type Response struct {
	Status 	string	`json:"ststus"`
	Error	string	`json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func Ok() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error: msg,
	}
}