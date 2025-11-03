package httpserver

func createStatusResponse(msg string) *ResponseForm {
	return &ResponseForm{Status: 200, Message: msg}
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

// CreateBucketForm example
type CreateBucketForm struct {
	BucketName string `json:"bucket_name" example:"test-bucket"`
}

// MoveFilesForm example
type MoveFilesForm struct {
	TargetDirectory string   `json:"location" example:"common-folder"`
	SourceDirectory string   `json:"src_folder_id" example:"unrecognized"`
	DocumentPaths   []string `json:"document_ids" example:"./indexer/watcher/test.txt"`
}

// RemoveFileForm example
type RemoveFileForm struct {
	FileName string `json:"file_name" example:"test-file.docx"`
}

// DownloadFileForm example
type DownloadFileForm struct {
	FileName string `json:"file_name" example:"test-file.docx"`
}

// ShareFileForm example
type ShareFileForm struct {
	FilePath    string `json:"file_path" example:"test-file.docx"`
	ExpiredSecs int32  `json:"expired_secs" example:"3600"`
}

// GetFilesForm example
type GetFilesForm struct {
	DirectoryName string `json:"directory" example:"test-folder/"`
}

// GetFileAttributesForm example
type GetFileAttributesForm struct {
	FilePath string `json:"file_path" example:"test-file.docx"`
}

// CopyFileForm example
type CopyFileForm struct {
	SrcPath string `json:"src_path" example:"old-test-document.docx"`
	DstPath string `json:"dst_path" example:"test-document.docx"`
}
