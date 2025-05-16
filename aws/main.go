package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var selectedProvider string
var envMap map[string]string


func main() {

	var secretsFile string

		secretsFile = "aws-secrets.txt"


	var err error
	envMap, err = godotenv.Read(secretsFile)
	if err != nil {
		log.Fatalf("Failed to load secrets from %s: %v", secretsFile, err)
	}
  AppAllowOrigins := envMap["APP_ALLOW_ORIGIN"]
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{AppAllowOrigins},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	err = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	var handler gin.HandlerFunc

		handler = func(c *gin.Context)
			uploadToAWS(c, envMap)



	r.POST("/upload", handler)

	appAddr := envMap["APP_ADDRESS"]
	if appAddr == "" {
		appAddr = "127.0.0.1"
	}

	appPort := envMap["APP_PORT"]
	if appPort == "" {
		appPort = "5000"
	}

	addr := appAddr + ":" + appPort

	log.Println("Listening on " + addr)
	log.Fatal(r.Run(addr))
}
func uploadToAWS(c *gin.Context, env map[string]string) {
	accessKey := env["OSS_ACCESS_KEY"]
	secretKey := env["OSS_SECRET_KEY"]
	region := env["OSS_REGION"]
	endpoint := env["OSS_ENDPOINT"]
	bucketName := env["OSS_BUCKET"]
	protocol := env["OSS_PROTOCOL"]
	if protocol == "" {
		protocol = "https"
	}
	username := c.PostForm("username")
	ip := c.PostForm("ip")
	username = strings.TrimSpace(username)
	if username != "" {
		username = strings.ReplaceAll(username, " ", "_") // optional sanitization
	}
	if accessKey == "" || secretKey == "" || region == "" || endpoint == "" || bucketName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing env variables"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type"})
		return
	}

	uuidStr := uuid.New().String()[:8]
	finalFilename := username + "-" + uuidStr + ext

	log.Printf("Upload received - username: %s, ip: %s, filename: %s", username, ip, finalFilename)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	hasher := md5.New()
	hasher.Write(buf.Bytes())
	md5sum := hex.EncodeToString(hasher.Sum(nil))

	reader := bytes.NewReader(buf.Bytes())

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               protocol + "://" + endpoint,
					SigningRegion:     region,
					HostnameImmutable: false,
				}, nil
			}),
		),
	)
	if err != nil {
		log.Printf("AWS config error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AWS config failed"})
		return
	}

	s3client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = false
	})

	_, err = s3client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        &bucketName,
		Key:           aws.String("images/" + finalFilename),
		Body:          reader,
		ContentLength: aws.Int64(int64(buf.Len())),
		ContentType:   aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		log.Printf("S3 upload failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "S3 upload failed"})
		return
	}

	stat := map[string]string{
		"filename":    finalFilename,
		"full_path":   "images/" + finalFilename,
		"md5":         md5sum,
		"ip":          ip,
		"upload_time": time.Now().Format("2006-01-02 15-04-05"),
	}

	statJSON, _ := json.MarshalIndent(stat, "", "  ")
	statReader := bytes.NewReader(statJSON)

	_, err = s3client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        &bucketName,
		Key:           aws.String("images_stat/" + strings.TrimSuffix(finalFilename, ext) + ".json"),
		Body:          statReader,
		ContentLength: aws.Int64(int64(statReader.Len())),
		ContentType:   aws.String("application/json"),
	})
	if err != nil {
		log.Printf("Failed to upload stat file: %v", err)
	}

	finalURL := protocol + "://" + bucketName + "." + endpoint + "/images/" + finalFilename
	log.Printf(finalURL)
	c.JSON(http.StatusOK, gin.H{"url": finalURL})
}
