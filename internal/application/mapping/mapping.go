package mapping

import (
	"watchtower/internal/application/dto"
	"watchtower/internal/domain/core/structures"
)

func FromDocument(doc *domain.Document) dto.StorageDocument {
	return dto.StorageDocument{
		Content:    doc.Content,
		SSDEEP:     doc.SSDEEP,
		Class:      doc.Class,
		FileName:   doc.FileName,
		FilePath:   doc.FilePath,
		FileSize:   doc.FileSize,
		CreatedAt:  doc.CreatedAt,
		ModifiedAt: doc.ModifiedAt,
		Tokens:     FromTokens(&doc.Tokens),
	}
}

func FromTokens(tokens *domain.Tokens) dto.ComputedTokens {
	return dto.ComputedTokens{
		ChunksCount: tokens.ChunksCount,
		ChunkedText: tokens.ChunkedText,
		Vectors:     tokens.Vectors,
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
