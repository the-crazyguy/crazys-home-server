package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"crzy-server/internal/environment"
	"crzy-server/internal/models/user"
)

var fileServerAddress string

func main() {
	environment.Load()

	fileServerAddress = os.Getenv("BACKEND_ADDRESS")

	router := gin.Default()
	// TODO: Max multipart memory

	router.StaticFile("/", "./public/ui/index.html")

	router.StaticFile("/login", "./public/login.html")
	router.POST("/login", postLogin)

	router.StaticFile("/logout", "./public/logout.html")
	router.POST("/logout", postLogout)

	router.StaticFile("/register", "./public/register.html")
	router.POST("/register", postRegister)

	// TODO: Auth middleware for GETting protected webpages?
	router.StaticFile("/upload", "./public/upload.html")
	router.POST("/upload", postUpload)

	router.GET("/download/:filename", getDownload)
	router.GET("/download/:filename/:owner", getDownload)

	router.Run("localhost:8081")
}

func getDownload(c *gin.Context) {
	owner := c.Param("owner")
	filename := c.Param("filename")

	requestURL := fileServerAddress + "/download"
	requestURL += "/" + filename
	if owner != "" {
		requestURL += "/" + owner
	}

	token, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Need to be logged in to upload files"})
		log.Println("postUpload: Need to be logged in to upload files")
		c.Abort()
		return
	}

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		log.Println("getDownload: Failed to create request")
		c.Abort()
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		log.Printf("getDownload: Failed to download file: %s\n", err.Error())
		c.Abort()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		log.Printf(
			"getDownload: Failed to download file. Got unexpected status code %d\n",
			resp.StatusCode,
		)
		c.Abort()
		return
	}

	c.Header("Content-Disposition", resp.Header.Get("Content-Disposition"))
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Status(http.StatusOK)

	// Copy the server's response
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while downloading file"})
		log.Printf("getDownload: Encountered an error while downloading file: %s\n", err.Error())
		c.Abort()
		return
	}
}

func postUpload(c *gin.Context) {
	token, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Need to be logged in to upload files"})
		log.Println("postUpload: Need to be logged in to upload files")
		c.Abort()
		return
	}

	// TODO: Send files 1 by 1? Performance comparison? Reliability? Broken connection?
	req, err := http.NewRequest("POST", fileServerAddress+"/form-upload", c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		log.Println("postUpload: Failed to create request")
		c.Abort()
		return
	}
	req.Header = c.Request.Header
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		log.Printf("postUpload: Failed to upload file: %s\n", err.Error())
		c.Abort()
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		log.Printf(
			"postUpload: Failed to upload file(s). Got unexpected status code %d\n",
			resp.StatusCode,
		)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File(s) uploaded successfully"})
}

func postRegister(c *gin.Context) {
	log.Println("postRegister entered")

	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid username and/or password"})
		log.Println("postRegister: Invalid username and/or password")
		c.Abort()
		return
	}

	au := user.AuthUser{Username: username, Password: password}

	jsonData, err := json.Marshal(au)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Failed to marshal registration data"},
		)
		log.Println("postLogin: Failed to marshal registration data")
		c.Abort()
		return
	}

	resp, err := http.Post(
		fileServerAddress+"/register",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform register request"})
		log.Println("postLogin: Failed to perform register request")
		c.Abort()
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": fmt.Sprintf("Remote server failed with code %d", resp.StatusCode)},
		)
		log.Printf("Remote server failed with code %d\n", resp.StatusCode)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful, please log in"})
}

func postLogout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "localhost", false, true)
	log.Println("User logged out")
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func postLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid username and/or password"})
		log.Println("postLogin: Invalid username and/or password")
		c.Abort()
		return
	}

	// NOTE: Hashing is done on the server side. It can also be done on the client
	// But for the sake of simplicity, it is not done. It should not be done *only*
	// on the client-side, however!
	au := user.AuthUser{Username: username, Password: password}

	jsonData, err := json.Marshal(au)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal login data"})
		log.Println("postLogin: Failed to marshal login data")
		c.Abort()
		return
	}

	resp, err := http.Post(
		fileServerAddress+"/login",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform login request"})
		log.Println("postLogin: Failed to perform login request")
		c.Abort()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": fmt.Sprintf("Remote server failed with code %d", resp.StatusCode)},
		)
		log.Printf("Remote server failed with code %d\n", resp.StatusCode)
		c.Abort()
		return
	}

	var responseData struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from server"})
		log.Println("postLogin: Invalid response from server")
		c.Abort()
		return
	}

	c.SetCookie("auth_token", responseData.Token, 12*3600, "/", "localhost", false, true)

	c.Redirect(http.StatusFound, "/")
}

// // TODO: Needed?
// func authnMiddleware(c *gin.Context) {
// 	cookie, err := c.Cookie("auth_token")
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Action requires logging in"})
// 		log.Println("authnMiddleware: Action requires logging in")
// 		c.Abort()
// 		return
// 	}
//
// 	c.Next()
// }
