package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"crzy-server/internal/models/user"
)

const fileServerAddress string = "http://localhost:8080"

func main() {
	fmt.Printf("Hello world!\n")
	router := gin.Default()
	// TODO: Max multipart memory

	router.StaticFile("/login", "./public/login.html")
	router.POST("/login", postLogin)

	router.Run("localhost:8081")
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

	// TODO: Store token
}
