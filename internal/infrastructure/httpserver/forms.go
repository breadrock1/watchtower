package httpserver

func createStatusResponse(status int, msg string) *ResponseForm {
	return &ResponseForm{Status: status, Message: msg}
}

// ResponseForm example
type ResponseForm struct {
	Status  int    `json:"status" example:"200"`
	Message string `json:"message" example:"Done"`
}

// BadRequestForm example
type BadRequestForm struct {
	Status  int    `json:"status" example:"400"`
	Message string `json:"message" example:"Bad Request message"`
}

// ServerErrorForm example
type ServerErrorForm struct {
	Status  int    `json:"status" example:"503"`
	Message string `json:"message" example:"Server Error message"`
}

// AddDirectoryToWatcherForm example
type AddDirectoryToWatcherForm struct {
	BucketName string `json:"bucket" example:"test-folder"`
	Suffix     string `json:"suffix" example:"./some-directory"`
}
