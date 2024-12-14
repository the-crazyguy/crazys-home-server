package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	// TODO: Max multipart memory

	// TODO: Add an environment variable/pass a variable for the on-system filepath
	// router.Static("/files", "./user-files")

	router.StaticFile("/", "./public/index.html")
	router.POST("/upload", postUpload)

	router.Run("localhost:8080")
}

// TODO: Make it so each separate file upload is a separate POST request
// so you can give each file a description. Would require the creation of a
// File storage/access api/db

func postUpload(c *gin.Context) {
	owner := c.PostForm("owner-name")
	desc := c.PostForm("description")

	// debug
	fmt.Println("Owner: ", owner)
	fmt.Println("Description: ", desc)

	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, "Form error: %s", err.Error())
		fmt.Println("Form error: ", err.Error())

		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.String(http.StatusBadRequest, "No files to upload.")
		fmt.Println("No files to upload")

		return
	}

	uploadDir := fmt.Sprintf("./user-files/%s", owner)

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.String(http.StatusInternalServerError, "Directory creation error: %s", err.Error())
		fmt.Println("Directory creation error: ", err.Error())
	}

	for _, f := range files {
		fileName := filepath.Base(f.Filename)

		uploadPath := fmt.Sprintf("%s/%s", uploadDir, fileName)

		if err := c.SaveUploadedFile(f, uploadPath); err != nil {
			c.String(http.StatusBadRequest, "Upload file error: %s", err.Error())
			fmt.Println("Upload file error: ", err.Error())

			return
		}
	}

	c.String(http.StatusOK, "Successfully uploaded %d files.", len(files))
}
