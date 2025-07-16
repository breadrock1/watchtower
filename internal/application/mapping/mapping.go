package mapping

import (
	"fmt"
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

func TaskStatusFromString(enumVal string) (dto.TaskStatus, error) {
	switch enumVal {
	case "received":
		return dto.Received, nil
	case "pending":
		return dto.Pending, nil
	case "processing":
		return dto.Processing, nil
	case "successful":
		return dto.Successful, nil
	case "failed":
		return dto.Failed, nil
	default:
		return dto.Pending, fmt.Errorf("unknown task status: %s", enumVal)
	}
}

func TaskStatusToInt(ts dto.TaskStatus) int {
	return int(ts)
}

func TaskStatusFromInt(enum int) dto.TaskStatus {
	return dto.TaskStatus(enum)
}
