package utils

import (
	"mime"
	"strings"
)

var (
	documentMimes = []string{
		"csv", "msword", "html", "json", "pdf",
		"rtf", "plain", "vnd.ms-excel", "xml",
		"vnd.ms-powerpoint", "vnd.oasis.opendocument.text",
		"vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"vnd.openxmlformats-officedocument.wordprocessingml.document",
		"vnd.openxmlformats-officedocument.presentationml.presentation",
	}
)

func ParseDocumentType(extension string) string {
	mimeType := mime.TypeByExtension(extension)
	attributes := strings.Split(mimeType, "/")
	switch attributes[0] {
	case "audio":
		return "audio"
	case "image":
		return "image"
	case "video":
		return "video"
	case "text":
		return "document"
	case "application":
		return extractApplicationMimeType(attributes[1])
	default:
		return "document"
	}
}

func extractApplicationMimeType(attribute string) string {
	for _, mimeType := range documentMimes {
		if mimeType == attribute {
			return "document"
		}
	}

	return "document"
}
