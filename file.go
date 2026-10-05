package woovi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"
)

// FilePurpose selects why a file is uploaded (dispute, limit, KYC).
type FilePurpose string

const (
	FilePurposeDisputeEvidence         FilePurpose = "DISPUTE_EVIDENCE"
	FilePurposeAccountLimitRequest     FilePurpose = "ACCOUNT_LIMIT_REQUEST"
	FilePurposeAccountRegisterDocument FilePurpose = "ACCOUNT_REGISTER_DOCUMENT"
)

// FilesService is POST /api/v1/files.
type FilesService struct {
	client *Client
}

// FileUploadParams uploads a PDF/image for a purpose; optional correlationID for idempotency.
type FileUploadParams struct {
	File     io.Reader // required
	FileName string    // required
	// ContentType is required. Allowed: application/pdf, image/png, image/jpeg, image/webp.
	ContentType string
	Purpose     FilePurpose // required
	// CorrelationID is optional. Same ID and purpose return the stored file.
	CorrelationID string
}

// File is an uploaded document with id and optional pre-signed URL.
type File struct {
	ID            string      `json:"id"`
	CorrelationID string      `json:"correlationID,omitempty"`
	Purpose       FilePurpose `json:"purpose,omitempty"`
	FileName      string      `json:"fileName,omitempty"`
	ContentType   string      `json:"contentType,omitempty"`
	Size          int64       `json:"size,omitempty"`
	URL           string      `json:"url,omitempty"`
	URLExpiresAt  time.Time   `json:"urlExpiresAt,omitempty"`
	CreatedAt     time.Time   `json:"createdAt,omitempty"`
}

type fileEnvelope struct {
	File File `json:"file"`
}

// Upload stores a file. Max size 10 MiB. Requires FILE_POST.
func (s *FilesService) Upload(ctx context.Context, params *FileUploadParams, opts ...RequestOption) (*File, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: FileUploadParams is required")
	}
	if params.File == nil {
		return nil, fmt.Errorf("woovi: file is required")
	}
	if strings.TrimSpace(params.FileName) == "" {
		return nil, fmt.Errorf("woovi: fileName is required")
	}
	if params.Purpose == "" {
		return nil, fmt.Errorf("woovi: purpose is required")
	}
	contentType := strings.TrimSpace(params.ContentType)
	if contentType == "" {
		contentType = guessContentType(params.FileName)
	}
	if contentType == "" {
		return nil, fmt.Errorf("woovi: contentType is required")
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, escapeQuotes(filepath.Base(params.FileName))))
	h.Set("Content-Type", contentType)
	part, err := w.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("woovi: create multipart file: %w", err)
	}
	if _, err := io.Copy(part, params.File); err != nil {
		return nil, fmt.Errorf("woovi: write multipart file: %w", err)
	}
	if err := w.WriteField("purpose", string(params.Purpose)); err != nil {
		return nil, fmt.Errorf("woovi: write purpose: %w", err)
	}
	if params.CorrelationID != "" {
		if err := w.WriteField("correlationID", params.CorrelationID); err != nil {
			return nil, fmt.Errorf("woovi: write correlationID: %w", err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("woovi: close multipart: %w", err)
	}

	var out fileEnvelope
	if err := s.client.doHTTP(ctx, http.MethodPost, "/api/v1/files", nil, w.FormDataContentType(), &buf, &out, opts...); err != nil {
		return nil, err
	}
	return &out.File, nil
}

func guessContentType(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
