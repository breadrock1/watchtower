package mapping

import (
	"watchtower/internal/application/dto"
	"watchtower/internal/domain/core/structures"
)

func FromDocument(doc *domain.Document) dto.StorageDocument {
	return dto.StorageDocument{
		FileName:   doc.FileName,
		FilePath:   doc.FilePath,
		FileSize:   doc.FileSize,
		Content:    doc.Content,
		CreatedAt:  doc.CreatedAt,
		ModifiedAt: doc.ModifiedAt,
	}
}

func TaskStatusFromString(enumVal string) dto.TaskStatus {
	switch enumVal {
	case "received":
		return dto.Received
	case "pending":
		return dto.Pending
	case "processing":
		return dto.Processing
	case "successful":
		return dto.Successful
	case "failed":
		return dto.Failed
	default:
		return dto.Pending
	}
}

func TaskStatusToInt(ts dto.TaskStatus) int {
	return int(ts)
}

func TaskStatusFromInt(enum int) dto.TaskStatus {
	return dto.TaskStatus(enum)
}
