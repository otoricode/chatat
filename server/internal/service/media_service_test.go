package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/pkg/apperror"
)

// --- Mock Media Repository ---
type mockMediaRepo struct {
	media map[uuid.UUID]*model.Media
}

func newMockMediaRepo() *mockMediaRepo {
	return &mockMediaRepo{media: make(map[uuid.UUID]*model.Media)}
}

func (m *mockMediaRepo) Create(_ context.Context, media *model.Media) error {
	m.media[media.ID] = media
	return nil
}

func (m *mockMediaRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Media, error) {
	media, ok := m.media[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return media, nil
}

func (m *mockMediaRepo) ListByContext(_ context.Context, contextType string, contextID uuid.UUID) ([]*model.Media, error) {
	var result []*model.Media
	for _, med := range m.media {
		if med.ContextType != nil && *med.ContextType == contextType &&
			med.ContextID != nil && *med.ContextID == contextID {
			result = append(result, med)
		}
	}
	return result, nil
}

func (m *mockMediaRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.media, id)
	return nil
}

// --- Mock Storage Service ---
type mockStorageService struct {
	files map[string][]byte
}

func newMockStorageService() *mockStorageService {
	return &mockStorageService{files: make(map[string][]byte)}
}

func (m *mockStorageService) Upload(_ context.Context, input UploadInput) (*UploadResult, error) {
	data, err := io.ReadAll(input.Data)
	if err != nil {
		return nil, err
	}
	m.files[input.Key] = data
	url := "http://localhost:9000/chatat-media/" + input.Key
	return &UploadResult{Key: input.Key, URL: url, Size: int64(len(data))}, nil
}

func (m *mockStorageService) GetURL(_ context.Context, key string) (string, error) {
	if _, ok := m.files[key]; !ok {
		return "", fmt.Errorf("not found")
	}
	return "http://localhost:9000/chatat-media/" + key + "?signed=1", nil
}

func (m *mockStorageService) Delete(_ context.Context, key string) error {
	delete(m.files, key)
	return nil
}

// --- Test Helper: create JPEG bytes ---
func createTestJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	return buf.Bytes()
}

// === Tests ===

func TestMediaService_UploadImage(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	imageSvc := NewImageService()
	svc := NewMediaService(mediaRepo, storageSvc, imageSvc)

	uploaderID := uuid.New()
	jpegData := createTestJPEG(800, 600)

	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uploaderID,
		Filename:    "photo.jpg",
		ContentType: "image/jpeg",
		Size:        int64(len(jpegData)),
		Data:        bytes.NewReader(jpegData),
		ContextType: "chat",
		ContextID:   uuid.New().String(),
	})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, model.MediaTypeImage, result.Type)
	assert.Equal(t, "photo.jpg", result.Filename)
	assert.NotEmpty(t, result.URL)
	assert.NotEmpty(t, result.ThumbnailURL)
	assert.NotNil(t, result.Width)
	assert.NotNil(t, result.Height)
	assert.Equal(t, 800, *result.Width)
	assert.Equal(t, 600, *result.Height)

	// Verify storage has 2 files (image + thumbnail)
	assert.Len(t, storageSvc.files, 2)

	// Verify media repo has record
	assert.Len(t, mediaRepo.media, 1)
}

func TestMediaService_UploadFile(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	imageSvc := NewImageService()
	svc := NewMediaService(mediaRepo, storageSvc, imageSvc)

	uploaderID := uuid.New()
	pdfData := []byte("%PDF-1.4 test content")

	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uploaderID,
		Filename:    "document.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(pdfData)),
		Data:        bytes.NewReader(pdfData),
	})

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, model.MediaTypeFile, result.Type)
	assert.Equal(t, "document.pdf", result.Filename)
	assert.NotEmpty(t, result.URL)
	assert.Empty(t, result.ThumbnailURL)
	assert.Nil(t, result.Width)
	assert.Nil(t, result.Height)

	// Only 1 file in storage (no thumbnail for files)
	assert.Len(t, storageSvc.files, 1)
}

func TestMediaService_UploadDisallowedType(t *testing.T) {
	svc := NewMediaService(newMockMediaRepo(), newMockStorageService(), NewImageService())

	_, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uuid.New(),
		Filename:    "script.sh",
		ContentType: "application/x-sh",
		Size:        100,
		Data:        bytes.NewReader([]byte("#!/bin/bash")),
	})

	require.Error(t, err)
	appErr, ok := err.(*apperror.AppError)
	require.True(t, ok)
	assert.Equal(t, "BAD_REQUEST", appErr.Code)
}

func TestMediaService_UploadImageTooLarge(t *testing.T) {
	svc := NewMediaService(newMockMediaRepo(), newMockStorageService(), NewImageService())

	_, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uuid.New(),
		Filename:    "huge.jpg",
		ContentType: "image/jpeg",
		Size:        17 * 1024 * 1024, // 17 MB > 16 MB limit
		Data:        bytes.NewReader([]byte{}),
	})

	require.Error(t, err)
	appErr, ok := err.(*apperror.AppError)
	require.True(t, ok)
	assert.Equal(t, "BAD_REQUEST", appErr.Code)
}

func TestMediaService_GetByID(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	imageSvc := NewImageService()
	svc := NewMediaService(mediaRepo, storageSvc, imageSvc)

	// Create media via upload
	uploaderID := uuid.New()
	pdfData := []byte("%PDF-1.4 test")

	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uploaderID,
		Filename:    "doc.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(pdfData)),
		Data:        bytes.NewReader(pdfData),
	})
	require.NoError(t, err)

	// Get by ID
	got, err := svc.GetByID(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, result.ID, got.ID)
	assert.Equal(t, "doc.pdf", got.Filename)
}

func TestMediaService_GetByID_NotFound(t *testing.T) {
	svc := NewMediaService(newMockMediaRepo(), newMockStorageService(), NewImageService())

	_, err := svc.GetByID(context.Background(), uuid.New())
	require.Error(t, err)
	appErr, ok := err.(*apperror.AppError)
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", appErr.Code)
}

func TestMediaService_Delete(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	svc := NewMediaService(mediaRepo, storageSvc, NewImageService())

	uploaderID := uuid.New()
	pdfData := []byte("test pdf")

	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uploaderID,
		Filename:    "del.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(pdfData)),
		Data:        bytes.NewReader(pdfData),
	})
	require.NoError(t, err)

	// Delete by owner
	err = svc.Delete(context.Background(), result.ID, uploaderID)
	require.NoError(t, err)

	// Storage should be empty
	assert.Len(t, storageSvc.files, 0)

	// DB should have 0 records
	assert.Len(t, mediaRepo.media, 0)
}

func TestMediaService_Delete_Forbidden(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	svc := NewMediaService(mediaRepo, storageSvc, NewImageService())

	uploaderID := uuid.New()
	pdfData := []byte("test")

	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uploaderID,
		Filename:    "secret.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(pdfData)),
		Data:        bytes.NewReader(pdfData),
	})
	require.NoError(t, err)

	// Try delete by different user
	otherUser := uuid.New()
	err = svc.Delete(context.Background(), result.ID, otherUser)
	require.Error(t, err)
	appErr, ok := err.(*apperror.AppError)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", appErr.Code)
}

func TestMediaService_GetDownloadURL(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	svc := NewMediaService(mediaRepo, storageSvc, NewImageService())

	pdfData := []byte("content")
	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uuid.New(),
		Filename:    "file.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(pdfData)),
		Data:        bytes.NewReader(pdfData),
	})
	require.NoError(t, err)

	url, err := svc.GetDownloadURL(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Contains(t, url, "signed=1")
}

func TestImageService_ProcessImage(t *testing.T) {
	svc := NewImageService()
	jpegData := createTestJPEG(2000, 1500) // larger than 1600px

	result, err := svc.ProcessImage(bytes.NewReader(jpegData))
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", result.ContentType)
	assert.True(t, result.Width <= 1600, "width should be <= 1600, got %d", result.Width)
	assert.True(t, result.Size > 0)
}

func TestImageService_ProcessImage_SmallImage(t *testing.T) {
	svc := NewImageService()
	jpegData := createTestJPEG(400, 300) // smaller than 1600px

	result, err := svc.ProcessImage(bytes.NewReader(jpegData))
	require.NoError(t, err)
	assert.Equal(t, 400, result.Width)
	assert.Equal(t, 300, result.Height)
}

func TestImageService_GenerateThumbnail(t *testing.T) {
	svc := NewImageService()
	jpegData := createTestJPEG(800, 600)

	result, err := svc.GenerateThumbnail(bytes.NewReader(jpegData), 0, 0)
	require.NoError(t, err)
	assert.True(t, result.Width <= 300 && result.Height <= 300,
		"thumbnail should fit in 300x300, got %dx%d", result.Width, result.Height)
	assert.Equal(t, "image/jpeg", result.ContentType)
}

func TestMediaService_UploadImageWithLargeResize(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	imageSvc := NewImageService()
	svc := NewMediaService(mediaRepo, storageSvc, imageSvc)

	// Create a 2000x1500 image (bigger than 1600 max)
	jpegData := createTestJPEG(2000, 1500)

	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uuid.New(),
		Filename:    "big.jpg",
		ContentType: "image/jpeg",
		Size:        int64(len(jpegData)),
		Data:        bytes.NewReader(jpegData),
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Width)
	assert.True(t, *result.Width <= 1600, "should be resized to <= 1600px wide")
}

// Verify context is properly stored
func TestMediaService_UploadWithContext(t *testing.T) {
	mediaRepo := newMockMediaRepo()
	storageSvc := newMockStorageService()
	svc := NewMediaService(mediaRepo, storageSvc, NewImageService())

	chatID := uuid.New()
	pdfData := []byte("context test")

	result, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uuid.New(),
		Filename:    "ctx.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(pdfData)),
		Data:        bytes.NewReader(pdfData),
		ContextType: "chat",
		ContextID:   chatID.String(),
	})
	require.NoError(t, err)

	// Verify the stored media has context
	media := mediaRepo.media[result.ID]
	require.NotNil(t, media.ContextType)
	assert.Equal(t, "chat", *media.ContextType)
	require.NotNil(t, media.ContextID)
	assert.Equal(t, chatID, *media.ContextID)
}

// Verify cleanup on DB error
func TestMediaService_UploadCleanupOnDBError(t *testing.T) {
	failingRepo := &failingMediaRepo{}
	storageSvc := newMockStorageService()
	svc := NewMediaService(failingRepo, storageSvc, NewImageService())

	pdfData := []byte("cleanup test")
	_, err := svc.Upload(context.Background(), MediaUploadInput{
		UploaderID:  uuid.New(),
		Filename:    "fail.pdf",
		ContentType: "application/pdf",
		Size:        int64(len(pdfData)),
		Data:        bytes.NewReader(pdfData),
	})
	require.Error(t, err)

	// File should be cleaned up from storage
	assert.Len(t, storageSvc.files, 0)
}

// --- Failing media repo for cleanup test ---
type failingMediaRepo struct{}

func (f *failingMediaRepo) Create(_ context.Context, _ *model.Media) error {
	return fmt.Errorf("db connection lost")
}
func (f *failingMediaRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Media, error) {
	return nil, fmt.Errorf("not found")
}
func (f *failingMediaRepo) ListByContext(_ context.Context, _ string, _ uuid.UUID) ([]*model.Media, error) {
	return nil, nil
}
func (f *failingMediaRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

// Suppress unused import warning
var _ = time.Now
