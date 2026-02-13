package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

// ImageService handles image processing: resize, compress, thumbnail.
type ImageService interface {
	ProcessImage(input io.Reader) (*ProcessedImage, error)
	GenerateThumbnail(input io.Reader, maxWidth, maxHeight int) (*ProcessedImage, error)
}

// ProcessedImage holds the result of image processing.
type ProcessedImage struct {
	Data        *bytes.Buffer
	Width       int
	Height      int
	Size        int64
	ContentType string // always "image/jpeg"
}

const (
	maxImageDimension     = 1600
	thumbnailMaxDimension = 300
	processedQuality      = 80
	thumbnailQuality      = 60
)

type imageService struct{}

// NewImageService creates a new image processing service.
func NewImageService() ImageService {
	return &imageService{}
}

func (s *imageService) ProcessImage(input io.Reader) (*ProcessedImage, error) {
	img, err := imaging.Decode(input, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Resize if larger than max dimension
	if w > maxImageDimension || h > maxImageDimension {
		img = imaging.Fit(img, maxImageDimension, maxImageDimension, imaging.Lanczos)
		bounds = img.Bounds()
		w, h = bounds.Dx(), bounds.Dy()
	}

	return encodeJPEG(img, w, h, processedQuality)
}

func (s *imageService) GenerateThumbnail(input io.Reader, maxWidth, maxHeight int) (*ProcessedImage, error) {
	if maxWidth <= 0 {
		maxWidth = thumbnailMaxDimension
	}
	if maxHeight <= 0 {
		maxHeight = thumbnailMaxDimension
	}

	img, err := imaging.Decode(input, imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decoding image for thumbnail: %w", err)
	}

	thumb := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
	bounds := thumb.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	return encodeJPEG(thumb, w, h, thumbnailQuality)
}

func encodeJPEG(img image.Image, w, h, quality int) (*ProcessedImage, error) {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encoding JPEG: %w", err)
	}

	return &ProcessedImage{
		Data:        buf,
		Width:       w,
		Height:      h,
		Size:        int64(buf.Len()),
		ContentType: "image/jpeg",
	}, nil
}
