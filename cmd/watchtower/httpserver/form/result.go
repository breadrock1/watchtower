package form

// Success example
type Success struct {
	Status  int    `json:"status" example:"200"`
	Message string `json:"message" example:"Done"`
}

func SuccessResponse(msg string) Success {
	return Success{
		Status:  200,
		Message: msg,
	}
}

// BadRequestError example
type BadRequestError struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"Bad Request message"`
}

// NotAcceptableError example
type NotAcceptableError struct {
	Status  int    `json:"status" example:"406"`
	Message string `json:"message" example:"Not acceptable request"`
}

// AuthError example
type AuthError struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"auth error"`
}

// NotFoundError example
type NotFoundError struct {
	Status  int    `json:"status" example:"404"`
	Message string `json:"message" example:"Not found"`
}

// InternalServerError example
type InternalServerError struct {
	Status  int    `json:"status" example:"500"`
	Message string `json:"message" example:"Internal server error message"`
}

// ServerUnavailableError example
type ServerUnavailableError struct {
	Status  int    `json:"status" example:"503"`
	Message string `json:"message" example:"Server unavailable error message"`
}
