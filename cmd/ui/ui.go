package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Printf("Hello world!\n")
	router := gin.Default()
	// TODO: Max multipart memory

	router.StaticFile("/", "./public/index.html")

	router.Run("localhost:8081")
}
