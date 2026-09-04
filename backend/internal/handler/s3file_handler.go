package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"uuid"

	"github.com/FTATM/iot-monitor-interface/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3File struct {
	url         string
	region      string
	accessKey   string
	secretKey   string
	imageBucket string
}

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func NewS3FileHandler(s3Config config.S3) *S3File {
	return &S3File{
		url:         s3Config.Url,
		region:      s3Config.Region,
		accessKey:   s3Config.AccessKey,
		secretKey:   s3Config.SecretKey,
		imageBucket: s3Config.ImageBucket,
	}
}

func (h *S3File) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	var res Response

	// 1. Limit max body size (e.g., 10MB) to protect against memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// 2. Parse form file
	file, _, err := r.FormFile("image")
	if err != nil {
		res.Message = "Invalid file or file size exceeds 10MB"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	defer file.Close()

	// 3. Read the first 512 bytes to sniff actual file content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		res.Message = "Failed to inspect file"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	// 4. Detect true MIME type from file header magic bytes
	detectedType := http.DetectContentType(buffer[:n])
	safeExt, ok := allowedImageTypes[detectedType]
	if !ok {
		res.Message = "Only image files (PNG, JPG, GIF, WEBP) are allowed"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// 5. Seek back to start so S3 reads the complete file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		res.Message = "Failed to process image stream"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	// 6. Use the verified extension rather than trusting the raw filename
	newFileName := uuid.New().String() + safeExt

	// 7. Push to Garage
	cfg, err := awsS3Config.LoadDefaultConfig(context.TODO(),
		awsS3Config.WithRegion(h.region),
		awsS3Config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(h.accessKey, h.secretKey, "")),
	)
	if err != nil {
		res.Message = "S3 config error"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(h.url)
		o.UsePathStyle = true
	})

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(h.imageBucket),
		Key:         aws.String(newFileName),
		Body:        file,
		ContentType: aws.String(detectedType), // Use detected MIME type
	})

	if err != nil {
		res.Message = "Failed to save image"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	publicUrl := fmt.Sprintf("/file/image/%s", newFileName)
	res.Data = map[string]any{
		"url": publicUrl,
	}
	respondJson(w, http.StatusOK, &res)
}

func (h *S3File) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	var res Response
	filename := r.PathValue("filename")
	if filename == "" {
		res.Message = "Invalid group Id"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	cfg, err := awsS3Config.LoadDefaultConfig(context.TODO(),
		awsS3Config.WithRegion(h.region),
		awsS3Config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(h.accessKey, h.secretKey, "")),
	)
	if err != nil {
		res.Message = "Error"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(h.url)
		o.UsePathStyle = true
	})

	out, err := s3Client.GetObject(r.Context(), &s3.GetObjectInput{
		Bucket: aws.String(h.imageBucket),
		Key:    aws.String(filename),
	})
	if err != nil {
		res.Message = "image not found"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	defer out.Body.Close()

	if out.ContentType != nil {
		w.Header().Set("Content-Type", *out.ContentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	io.Copy(w, out.Body)
}
