package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	router := gin.Default()
	// TODO: Max multipart memory

	// TODO: Add an environment variable/pass a variable for the on-system filepath
	// router.Static("/files", "./user-files")

	// TODO: Use templates
	router.StaticFile("/", "./public/index.html")
	router.StaticFile("/login", "./public/login.html")

	router.POST("/login-form", postLoginForm)
	router.POST("/upload", postUpload)
	router.GET("/download/:filename", authnMiddlewareMock, getDownload)

	router.Run("localhost:8080")
}

func authnMiddlewareMock(c *gin.Context) {
	c.Set("username", "crzy")
	c.Next()
}

func authnMiddleware(c *gin.Context) {
	// Step 1: Get the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		log.Println("Error: No authorization header")
		c.Abort()
		return
	}

	// Step 2: Extract the token string
	tknStr := strings.Split(authHeader, " ")[1]
	// TODO: REMOVE, here for testing
	log.Println(tknStr)
	// Step 3: Parse the token (i.e. with jwt)

	// Step 4: Check for token validity

	// Step 5: Extract useful information from claims i.e. Username

	// Step 6: Set the extracted information in the context (c)

	// Step 7: Continue to next handler
	c.Next()
}

func getDownload(c *gin.Context) {
	// Step 1: Get username from context
	username, exists := c.Get("username")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username missing from context"})
		log.Println("Error: Username missing from context")
		c.Abort()
		return
	}
	// Step 2: Get filename from URL (or rework to a POST request w/ JSON?)
	fileName := c.Param("filename")

	if fileName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "filename is empty"})
		log.Println("Error: filename missing from url")
		c.Abort()
		return
	}

	// Step 3: Construct filepath
	path := filepath.Join("user-files", username.(string), filepath.Clean(fileName))

	// WARN: Test for path traversal!
	if !strings.HasPrefix(path, "user-files") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		log.Printf("Error: access to %s is denied\n", path)
		c.Abort()
		return
	}

	// Step 4: Check if the path exists
	if _, err := os.Stat(path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		log.Printf("Error: file %s not found\n", path)
		c.Abort()
		return
	}

	// Step 5: Serve the file
	// TODO: Test if this supports range requests
	// TODO: Test with large files
	c.FileAttachment(path, fileName)

	// Alternatively: Serve via http.SetveContent to handle range requests if c.FileAttachment doesn't work
	// http.ServeContent(c.Writer, c.Request, fi)
}

func postLogin(c *gin.Context) {
	// Step 1: Get params from provided json (username, password)
	// Step 2: Validate username and password
	// Step 3: Generate JWT/Bearer token
	// Step 4: Return bearer token as a response
}

func postLoginForm(c *gin.Context) {
	var u struct {
		Username string `form:"username" binding:"required"`
	}
	if err := c.ShouldBind(&u); err != nil {
		c.String(http.StatusBadRequest, "Binding error: %s", err.Error())
		log.Println("Binding error: ", err.Error())

		return
	}

	log.Println("Username: ", u.Username)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": u})
}

// TODO: Make it so each separate file upload is a separate POST request
// so you can give each file a description. Would require the creation of a
// File storage/access api/db

func postUpload(c *gin.Context) {
	owner := c.PostForm("owner-name")
	desc := c.PostForm("description")

	// debug
	log.Println("Owner: ", owner)
	log.Println("Description: ", desc)

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Form error: %s", err.Error())
		log.Println("Form error: ", err.Error())

		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.String(http.StatusBadRequest, "No files to upload.")
		log.Println("No files to upload")

		return
	}

	uploadDir := filepath.Join(".", "user-files", owner)
	log.Printf("uploadDir: %s\n", uploadDir)

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.String(http.StatusInternalServerError, "Directory creation error: %s", err.Error())
		log.Println("Directory creation error: ", err.Error())
	}

	for _, f := range files {
		fileName := filepath.Base(f.Filename)

		uploadPath := filepath.Join(uploadDir, fileName)
		log.Printf("uploadPath: %s\n", uploadPath)

		if err := c.SaveUploadedFile(f, uploadPath); err != nil {
			c.String(http.StatusBadRequest, "Upload file error: %s", err.Error())
			log.Println("Upload file error: ", err.Error())

			return
		}
	}

	c.String(http.StatusOK, "Successfully uploaded %d files.", len(files))
}
