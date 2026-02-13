# Phase 11: Media System

> Implementasi upload, download, dan preview media (foto, file).
> Phase ini menghasilkan kemampuan berbagi file dan gambar di chat.

**Estimasi:** 4 hari
**Dependency:** Phase 07 (Chat Personal)
**Output:** Media upload/download, image compression, preview, gallery.

---

## Task 11.1: Storage Backend (S3-compatible)

**Input:** Go server dari Phase 01
**Output:** File storage service menggunakan S3-compatible storage

### Steps:
1. Install dependency:
   ```bash
   go get github.com/aws/aws-sdk-go-v2
   go get github.com/aws/aws-sdk-go-v2/service/s3
   ```
2. Buat `internal/service/storage_service.go`:
   ```go
   type StorageService interface {
       Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
       GetURL(ctx context.Context, key string) (string, error)
       Delete(ctx context.Context, key string) error
       GetPresignedUploadURL(ctx context.Context, key string, contentType string) (string, error)
   }

   type UploadInput struct {
       Data        io.Reader
       Key         string // e.g., "media/chat/{chatId}/{uuid}.jpg"
       ContentType string
       Size        int64
   }

   type UploadResult struct {
       Key  string `json:"key"`
       URL  string `json:"url"`
       Size int64  `json:"size"`
   }
   ```
3. Implementasi:
   - Development: MinIO (Docker container)
   - Production: AWS S3 atau compatible (DigitalOcean Spaces, Backblaze B2)
4. Bucket structure:
   ```
   chatat-media/
   ├── avatars/{userId}/avatar.jpg
   ├── media/chat/{chatId}/{uuid}.{ext}
   ├── media/topic/{topicId}/{uuid}.{ext}
   ├── media/document/{docId}/{uuid}.{ext}
   └── backups/{userId}/{timestamp}.enc
   ```
5. Add MinIO to docker-compose.yml:
   ```yaml
   minio:
     image: minio/minio:latest
     command: server /data --console-address ":9001"
     environment:
       MINIO_ROOT_USER: minioadmin
       MINIO_ROOT_PASSWORD: minioadmin
     ports:
       - "9000:9000"
       - "9001:9001"
     volumes:
       - miniodata:/data
   ```
6. Config:
   ```env
   S3_ENDPOINT=http://localhost:9000
   S3_BUCKET=chatat-media
   S3_ACCESS_KEY=minioadmin
   S3_SECRET_KEY=minioadmin
   S3_REGION=us-east-1
   ```

### Acceptance Criteria:
- [x] MinIO running in Docker
- [x] Upload file to S3
- [x] Get signed URL for download
- [x] Delete file
- [x] Presigned upload URL for direct client upload
- [x] Bucket and key structure organized

### Testing:
- [x] Integration test: upload file
- [x] Integration test: get URL
- [x] Integration test: delete file
- [x] Integration test: presigned URL

---

## Task 11.2: Image Processing

**Input:** Task 11.1
**Output:** Image compression dan thumbnail generation

### Steps:
1. Install dependency:
   ```bash
   go get github.com/disintegration/imaging
   ```
2. Buat `internal/service/image_service.go`:
   ```go
   type ImageService interface {
       ProcessImage(input io.Reader) (*ProcessedImage, error)
       GenerateThumbnail(input io.Reader, maxWidth, maxHeight int) (*ProcessedImage, error)
   }

   type ProcessedImage struct {
       Data        io.Reader
       Width       int
       Height      int
       Size        int64
       ContentType string    // always "image/jpeg"
   }
   ```
3. Implementasi ProcessImage:
   - Decode input (JPEG, PNG, WebP, HEIC)
   - Resize if larger than 1600px on longest side
   - Compress to JPEG quality 80
   - Strip EXIF metadata (privacy)
   - Return processed image
4. Implementasi GenerateThumbnail:
   - Resize to max 300x300 (maintaining aspect ratio)
   - Compress to JPEG quality 60
   - Return thumbnail
5. Processing pipeline:
   - Original → process → upload as `{key}.jpg`
   - Original → thumbnail → upload as `{key}_thumb.jpg`

### Acceptance Criteria:
- [x] Image resized to max 1600px
- [x] EXIF metadata stripped
- [x] Thumbnail generated (300px)
- [x] JPEG compression applied
- [x] Supports JPEG, PNG, WebP input
- [x] Output always JPEG

### Testing:
- [x] Unit test: process large image
- [x] Unit test: process small image (no resize)
- [x] Unit test: thumbnail generation
- [x] Unit test: EXIF stripping
- [x] Benchmark: processing performance

---

## Task 11.3: Media Upload API

**Input:** Task 11.1, 11.2
**Output:** Media upload endpoints

### Steps:
1. Buat migration `000008_media.up.sql`:
   ```sql
   CREATE TABLE media (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     uploader_id UUID NOT NULL REFERENCES users(id),
     type VARCHAR(10) NOT NULL CHECK(type IN ('image', 'file')),
     filename VARCHAR(255) NOT NULL,
     content_type VARCHAR(100) NOT NULL,
     size INTEGER NOT NULL,
     width INTEGER,
     height INTEGER,
     storage_key VARCHAR(500) NOT NULL,
     thumbnail_key VARCHAR(500),
     context_type VARCHAR(20) CHECK(context_type IN ('chat', 'topic', 'document')),
     context_id UUID,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_media_uploader_id ON media(uploader_id);
   CREATE INDEX idx_media_context ON media(context_type, context_id);
   ```
2. Buat `internal/handler/media_handler.go`:
   - `POST /api/v1/media/upload` → upload file
     - Multipart form data: file + metadata
     - Max file size: 16MB (images), 100MB (files)
     - Process image (if image) → generate thumbnail
     - Upload to S3
     - Create media record in DB
     - Response: `{"data": {"id": "uuid", "url": "...", "thumbnailUrl": "...", "size": 12345}}`
   - `GET /api/v1/media/:mediaId` → get media info
   - `GET /api/v1/media/:mediaId/download` → redirect to signed URL
3. File type validation:
   - Images: JPEG, PNG, WebP, HEIC, GIF
   - Files: PDF, DOC/DOCX, XLS/XLSX, PPT/PPTX, TXT, ZIP
   - Reject: executables, scripts
4. Media message integration:
   - When sending image message:
     1. Upload media → get media ID
     2. Send message with type="image" and metadata=`{"mediaId": "uuid"}`

### Acceptance Criteria:
- [x] Multipart upload berfungsi
- [x] File size limits enforced
- [x] File type validation
- [x] Image processing pipeline
- [x] Thumbnail generated for images
- [x] Media record stored in DB
- [x] Signed URL for download

### Testing:
- [x] Integration test: upload image
- [x] Integration test: upload file
- [x] Integration test: file too large → error
- [x] Integration test: invalid file type → error
- [x] Integration test: download via signed URL

---

## Task 11.4: Media UI (Frontend)

**Input:** Task 11.3, Chat screen dari Phase 07
**Output:** Image picker, camera, preview, gallery

### Steps:
1. Install dependencies:
   ```bash
   npm install react-native-image-picker
   npm install react-native-fast-image   # or expo-image
   npm install react-native-image-zoom-viewer
   ```
2. Buat `src/components/chat/AttachmentPicker.tsx`:
   - Bottom sheet with options:
     - 📷 Kamera: open camera → take photo → upload
     - 🖼️ Galeri: open image picker → select → upload
     - 📎 File: open document picker → select → upload
     - 📄 Dokumen: create new document (Phase 13)
3. Buat `src/components/chat/ImageMessage.tsx`:
   - Thumbnail in chat bubble (max 250px wide)
   - Tap → full-screen image viewer
   - Loading placeholder while downloading
   - Failed state with retry
4. Buat `src/components/chat/FileMessage.tsx`:
   - File card in chat bubble:
     - File icon (based on type)
     - Filename
     - File size
   - Tap → download + open with system viewer
5. Buat `src/screens/chat/ImageViewerScreen.tsx`:
   - Full-screen image viewer
   - Pinch-to-zoom
   - Swipe to dismiss
   - Share/save buttons
6. Buat `src/services/api/media.ts`:
   ```tsx
   export const mediaApi = {
     upload: (file: FormData) => apiClient.post('/media/upload', file, {
       headers: { 'Content-Type': 'multipart/form-data' },
       onUploadProgress: (progress) => { ... },
     }),
     getInfo: (mediaId: string) => apiClient.get(`/media/${mediaId}`),
   };
   ```
7. Upload progress indicator in chat bubble

### Acceptance Criteria:
- [x] Image picker: camera + gallery
- [x] File picker: document types
- [x] Upload progress visible
- [x] Thumbnail in chat bubble
- [x] Full-screen image viewer with zoom
- [x] File download + open
- [x] Loading/error states
- [x] Image caching (FastImage)

### Testing:
- [x] Component test: AttachmentPicker options
- [x] Component test: ImageMessage rendering
- [x] Component test: FileMessage rendering
- [x] Component test: upload progress
- [x] Integration test: upload → display in chat

---

## Phase 11 Review

### Testing Checklist:
- [x] Backend: S3 upload/download
- [x] Backend: image processing + thumbnail
- [x] Backend: file type validation
- [x] Backend: file size limits
- [x] Frontend: image picker (camera + gallery)
- [x] Frontend: file picker
- [x] Frontend: upload with progress
- [x] Frontend: image in chat bubble + full viewer
- [x] Frontend: file download
- [x] End-to-end: send image → receiver sees it
- [x] `go test ./...` + `npm test` pass

### Review Checklist:
- [x] Media types sesuai spec requirements
- [x] EXIF stripped for privacy
- [x] File size limits reasonable
- [x] Image quality/compression balanced
- [x] Commit: `feat(media): implement media upload and sharing`
