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
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var selectedProvider string
var envMap map[string]string


func main() {
	var secretsFile string
	secretsFile = "aliyun-secrets.txt"
	var err error
	envMap, err = godotenv.Read(secretsFile)
	if err != nil {
		log.Fatalf("Failed to load secrets from %s: %v", secretsFile, err)
	}
	AppAllowOrigins := envMap["APP_ALLOW_ORIGIN"]
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{ AppAllowOrigins },
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
  uploadToAliyun(c, envMap)

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


func uploadToAliyun(c *gin.Context, env map[string]string) {
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
		username += "-"
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
	finalFilename := username + uuidStr + ext

	log.Printf("Upload received - username: %s, ip: %s, filename: %s", username, ip, finalFilename)

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey)).
		WithRegion(region).
		WithEndpoint(endpoint)

	client := oss.NewClient(cfg)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	
	hasher := md5.New()
	_, err = hasher.Write(buf.Bytes())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute hash"})
		return
	}
	md5sum := hex.EncodeToString(hasher.Sum(nil))

	reader := bytes.NewReader(buf.Bytes())

	req := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr("images/" + finalFilename),
		Body:   reader,
	}

	_, err = client.PutObject(context.TODO(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed: " + err.Error()})
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
	statUploadReq := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr("images_stat/" + strings.TrimSuffix(finalFilename, ext) + ".json"),
		Body:   statReader,
	}
	_, err = client.PutObject(context.TODO(), statUploadReq)
	if err != nil {
		log.Printf("Failed to upload stat file: %v", err)
	}

	cleanEndpoint := strings.Replace(endpoint, "-internal", "", 1)
	publicEndpoint := cleanEndpoint

	finalURL := protocol + "://" + bucketName + "." + publicEndpoint + "/images/" + finalFilename
	c.JSON(http.StatusOK, gin.H{"url": finalURL})
}
