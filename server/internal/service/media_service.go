package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/pkg/apperror"
)

// MediaService handles media upload, download, and management.
type MediaService interface {
	Upload(ctx context.Context, input MediaUploadInput) (*model.MediaResponse, error)
	GetByID(ctx context.Context, mediaID uuid.UUID) (*model.MediaResponse, error)
	GetDownloadURL(ctx context.Context, mediaID uuid.UUID) (string, error)
	Delete(ctx context.Context, mediaID, userID uuid.UUID) error
}

// MediaUploadInput holds parameters for uploading media.
type MediaUploadInput struct {
	UploaderID  uuid.UUID
	Filename    string
	ContentType string
	Size        int64
	Data        io.Reader
	ContextType string // "chat", "topic", "document", or empty
	ContextID   string // UUID string or empty
}

// Allowed content types.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/heic": true,
	"image/gif":  true,
}

var allowedFileTypes = map[string]bool{
	"application/pdf":               true,
	"application/msword":            true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.ms-excel":      true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.ms-powerpoint": true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":                    true,
	"application/zip":               true,
	"application/x-zip-compressed":  true,
}

const (
	maxImageSize = 16 * 1024 * 1024  // 16 MB
	maxFileSize  = 100 * 1024 * 1024 // 100 MB
)

type mediaService struct {
	mediaRepo   repository.MediaRepository
	storageSvc  StorageService
	imageSvc    ImageService
}

// NewMediaService creates a new media service.
func NewMediaService(mediaRepo repository.MediaRepository, storageSvc StorageService, imageSvc ImageService) MediaService {
	return &mediaService{
		mediaRepo:  mediaRepo,
		storageSvc: storageSvc,
		imageSvc:   imageSvc,
	}
}

func (s *mediaService) Upload(ctx context.Context, input MediaUploadInput) (*model.MediaResponse, error) {
	isImage := allowedImageTypes[input.ContentType]
	isFile := allowedFileTypes[input.ContentType]

	if !isImage && !isFile {
		return nil, apperror.BadRequest("tipe file tidak diizinkan: " + input.ContentType)
	}

	if isImage && input.Size > maxImageSize {
		return nil, apperror.BadRequest("ukuran gambar maksimal 16MB")
	}
	if isFile && input.Size > maxFileSize {
		return nil, apperror.BadRequest("ukuran file maksimal 100MB")
	}

	mediaID := uuid.New()
	mediaType := model.MediaTypeFile
	if isImage {
		mediaType = model.MediaTypeImage
	}

	// Determine storage key path
	contextPath := "general"
	if input.ContextType != "" && input.ContextID != "" {
		contextPath = fmt.Sprintf("%s/%s", input.ContextType, input.ContextID)
	}

	ext := strings.ToLower(filepath.Ext(input.Filename))
	if ext == "" {
		ext = ".bin"
	}

	var (
		storageKey   string
		thumbnailKey *string
		width        *int
		height       *int
	)

	if isImage {
		// Read data to buffer for processing
		data, err := io.ReadAll(input.Data)
		if err != nil {
			return nil, fmt.Errorf("reading upload data: %w", err)
		}

		// Process image (resize, compress, strip EXIF)
		processed, err := s.imageSvc.ProcessImage(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("processing image: %w", err)
		}

		storageKey = fmt.Sprintf("media/%s/%s.jpg", contextPath, mediaID.String())

		// Upload processed image
		_, err = s.storageSvc.Upload(ctx, UploadInput{
			Data:        bytes.NewReader(processed.Data.Bytes()),
			Key:         storageKey,
			ContentType: processed.ContentType,
			Size:        processed.Size,
		})
		if err != nil {
			return nil, fmt.Errorf("uploading processed image: %w", err)
		}

		// Generate and upload thumbnail
		thumb, err := s.imageSvc.GenerateThumbnail(bytes.NewReader(data), 0, 0)
		if err == nil {
			thumbKey := fmt.Sprintf("media/%s/%s_thumb.jpg", contextPath, mediaID.String())
			_, uploadErr := s.storageSvc.Upload(ctx, UploadInput{
				Data:        bytes.NewReader(thumb.Data.Bytes()),
				Key:         thumbKey,
				ContentType: thumb.ContentType,
				Size:        thumb.Size,
			})
			if uploadErr == nil {
				thumbnailKey = &thumbKey
			}
		}

		w := processed.Width
		h := processed.Height
		width = &w
		height = &h
		input.Size = processed.Size
	} else {
		storageKey = fmt.Sprintf("media/%s/%s%s", contextPath, mediaID.String(), ext)

		_, err := s.storageSvc.Upload(ctx, UploadInput{
			Data:        input.Data,
			Key:         storageKey,
			ContentType: input.ContentType,
			Size:        input.Size,
		})
		if err != nil {
			return nil, fmt.Errorf("uploading file: %w", err)
		}
	}

	// Parse optional context
	var contextType *string
	var contextID *uuid.UUID
	if input.ContextType != "" && input.ContextID != "" {
		ct := input.ContextType
		contextType = &ct
		cid, err := uuid.Parse(input.ContextID)
		if err == nil {
			contextID = &cid
		}
	}

	media := &model.Media{
		ID:           mediaID,
		UploaderID:   input.UploaderID,
		Type:         mediaType,
		Filename:     input.Filename,
		ContentType:  input.ContentType,
		Size:         int(input.Size),
		Width:        width,
		Height:       height,
		StorageKey:   storageKey,
		ThumbnailKey: thumbnailKey,
		ContextType:  contextType,
		ContextID:    contextID,
		CreatedAt:    time.Now(),
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		// Cleanup uploaded files on DB error
		_ = s.storageSvc.Delete(ctx, storageKey)
		if thumbnailKey != nil {
			_ = s.storageSvc.Delete(ctx, *thumbnailKey)
		}
		return nil, fmt.Errorf("saving media record: %w", err)
	}

	return s.toResponse(ctx, media)
}

func (s *mediaService) GetByID(ctx context.Context, mediaID uuid.UUID) (*model.MediaResponse, error) {
	media, err := s.mediaRepo.FindByID(ctx, mediaID)
	if err != nil {
		return nil, apperror.NotFound("media", mediaID.String())
	}
	return s.toResponse(ctx, media)
}

func (s *mediaService) GetDownloadURL(ctx context.Context, mediaID uuid.UUID) (string, error) {
	media, err := s.mediaRepo.FindByID(ctx, mediaID)
	if err != nil {
		return "", apperror.NotFound("media", mediaID.String())
	}
	return s.storageSvc.GetURL(ctx, media.StorageKey)
}

func (s *mediaService) Delete(ctx context.Context, mediaID, userID uuid.UUID) error {
	media, err := s.mediaRepo.FindByID(ctx, mediaID)
	if err != nil {
		return apperror.NotFound("media", mediaID.String())
	}

	if media.UploaderID != userID {
		return apperror.Forbidden("hanya uploader yang dapat menghapus media")
	}

	// Delete from storage
	_ = s.storageSvc.Delete(ctx, media.StorageKey)
	if media.ThumbnailKey != nil {
		_ = s.storageSvc.Delete(ctx, *media.ThumbnailKey)
	}

	return s.mediaRepo.Delete(ctx, mediaID)
}

func (s *mediaService) toResponse(ctx context.Context, media *model.Media) (*model.MediaResponse, error) {
	url, err := s.storageSvc.GetURL(ctx, media.StorageKey)
	if err != nil {
		return nil, err
	}

	var thumbnailURL string
	if media.ThumbnailKey != nil {
		thumbURL, err := s.storageSvc.GetURL(ctx, *media.ThumbnailKey)
		if err == nil {
			thumbnailURL = thumbURL
		}
	}

	return &model.MediaResponse{
		ID:           media.ID,
		Type:         media.Type,
		Filename:     media.Filename,
		ContentType:  media.ContentType,
		Size:         media.Size,
		Width:        media.Width,
		Height:       media.Height,
		URL:          url,
		ThumbnailURL: thumbnailURL,
		CreatedAt:    media.CreatedAt,
	}, nil
}
