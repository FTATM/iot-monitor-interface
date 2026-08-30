package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/FTATM/iot-monitor-interface/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3File struct {
	url         string
	region      string
	accessKey   string
	secretKey   string
	imageBucket string
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
	cfg, err := awsS3Config.LoadDefaultConfig(context.TODO(),
		awsS3Config.WithRegion(h.region),
		awsS3Config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(h.accessKey, h.secretKey, "")),
	)

	// 2. Create the S3 client, overriding the endpoint and forcing Path-Style
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// ⚡ USE THE DOCKER SERVICE NAME, NOT LOCALHOST
		o.BaseEndpoint = aws.String(h.url)
		// o.BaseEndpoint = aws.String("http://garage:3900")
		o.UsePathStyle = true
	})

	// 3. Get the file from Vue
	file, header, err := r.FormFile("image")
	if err != nil {
		res.Message = "Failed to read image"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	newFileName := uuid.New().String() + ext

	// 4. Push to Garage
	bucketName := h.imageBucket

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(newFileName),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
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
