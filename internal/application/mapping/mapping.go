package mapping

import (
	"watchtower/internal/application/dto"
	"watchtower/internal/domain/core/structures"
)

func FromDocument(doc *domain.Document) dto.StorageDocument {
	return dto.StorageDocument{
		Content:    doc.Content,
		SSDEEP:     doc.SSDEEP,
		ID:         doc.ID,
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

func TaskStatusFromString(enumVal string) domain.TaskStatus {
	switch enumVal {
	case "received":
		return domain.Received
	case "pending":
		return domain.Pending
	case "processing":
		return domain.Processing
	case "successful":
		return domain.Successful
	case "failed":
		return domain.Failed
	default:
		return domain.Pending
	}
}

func TaskStatusToInt(ts domain.TaskStatus) int {
	return int(ts)
}

func TaskStatusFromInt(enum int) domain.TaskStatus {
	return domain.TaskStatus(enum)
}
