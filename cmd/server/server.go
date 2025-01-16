package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"crzy-server/internal/authentication"
	usermodel "crzy-server/internal/models/user"
	"crzy-server/internal/repo/user"
)

var userRepo user.UserRepository

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// TODO: Define in .env
	dbCfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Unable to parse DATABASE_URL:", err.Error())
		return
	}

	userRepo, err = user.NewPostgresUserRepository(dbCfg)
	if err != nil {
		log.Println("Failed to create a repository: ", err.Error())
		return
	}

	router := gin.Default()
	// secure := router.Group("/secure", authnMiddleware)
	// TODO: Remove once done w/ testing
	// unsecure := router.Group("/unsecure", authnMiddlewareMock)
	// TODO: Max multipart memory

	// TODO: Add an environment variable/pass a variable for the on-system filepath
	// router.Static("/files", "./user-files")

	// TODO: Remove, here for quick testing
	router.StaticFile("/", "./public/index.html")
	// router.StaticFile("/login", "./public/login.html")

	// router.POST("/login-form", postLoginForm)
	router.POST("/register", postRegister)
	router.POST("/login", postLogin)

	// TODO: Rename/rework
	router.POST("/form-upload", postFormUpload)

	router.GET("/download/:filename", authnMiddleware, getDownload)
	router.GET("/download/:filename/:owner", authnMiddleware, getDownload)

	// router.GET("/download/:filename", authnMiddleware, getDownload)
	// unsecure.GET("/download/:filename", getDownload)

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
	token, err := authentication.ParseToken(tknStr)
	// Step 4: Check for token validity
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not parse auth token"})
		log.Println("Error: Could not parse auth token: ", err.Error())
		c.Abort()
		return
	}

	if !token.Valid {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token"})
		log.Println("Error: Invalid token")
		c.Abort()
		return
	}

	// Step 5: Extract useful information from claims i.e. Username
	claims, err := authentication.GetClaims(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not parse claims"})
		log.Println("Error: Could not parse claims: ", err.Error())
		c.Abort()
		return
	}

	// Step 6: Set the extracted information in the context (c)
	// NOTE: Is there a better/more secure way to pass it? Can it be tampered with?
	c.Set("username", claims.Username)

	// Step 7: Continue to next handler
	c.Next()
}

func getDownload(c *gin.Context) {
	// Step 1: Get username from context
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "username is empty or not found"})
		log.Println("Error: username is empty or not found")
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

	ownerUsername := c.Param("owner")
	if ownerUsername == "" {
		// If no owner is provided, assume the file belongs to the logged-in user
		ownerUsername = username
	} else if ownerUsername == username {
		// User is attempting to access their own files, OK
	} else {
		// Owner passed, check if owner trusts the currently logged-in user
		owner, err := userRepo.GetByUsername(ownerUsername)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			log.Printf("getDownload: Could not find user %q", ownerUsername)
			c.Abort()
			return
		}

		currentUser, err := userRepo.GetByUsername(username)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			log.Printf("getDownload: Error looking up user %q in the database", username)
			c.Abort()
			return
		}

		canAccess, err := userRepo.Trusts(owner.ID, currentUser.ID)
		if err != nil {
			// Message vague on purpose
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			log.Printf("getDownload: Failed to determine trust relationship: %s", err.Error())
			c.Abort()
			return
		}

		if !canAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient access rights"})
			log.Printf("getDownload: user %q isn't trusted %q", username, ownerUsername)
			c.Abort()
			return
		}
	}
	// Sufficient Access

	// Step 3: Construct filepath
	path := filepath.Join("user-files", ownerUsername, filepath.Clean(fileName))

	// WARN: Test for path traversal!
	if !strings.HasPrefix(path, "user-files") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		log.Printf("Error: access to %s is denied\n", path)
		c.Abort()
		return
	}

	// Step 4: Check if the path exists
	if _, err := os.Stat(path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		log.Printf("getDownload: file %q not found\n", path)
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

// TODO: (BIG) Register with email to avoid disclosing if the user is found
func postRegister(c *gin.Context) {
	// Step 1: Get params from provided json (username, password)
	// NOTE: Can use another type or just user.User to accomodate for more info
	var userAuth usermodel.AuthUser
	if err := c.ShouldBindJSON(&userAuth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		log.Printf("postRegister: cannot bind to AuthUser: %s", err.Error())
		c.Abort()
		return
	}

	// Step 2: Check if user already exists (username/email)
	_, err := userRepo.GetByUsername(userAuth.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		// User exists, we have to disclose it unfortunately
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		log.Printf("postRegister: User already exists: %s", err.Error())
		c.Abort()
		return
	}

	// Step 3: Hash user's password
	hashedPassword, err := authentication.HashPassword(userAuth.Password)
	if err != nil {
		// Error message vague on purpose
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		log.Printf("postRegister: cannot hash password: %s", err.Error())
		c.Abort()
		return
	}

	// Step 4: Store user information
	user := &usermodel.User{
		Username:  userAuth.Username,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
	}

	if err := userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		log.Printf("postRegister: Could not create user in database: %s", err.Error())
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "User created"})
	log.Printf("user '%s' created", user.Username)
}

func postLogin(c *gin.Context) {
	// Step 1: Get params from provided json (username, password)
	var userAuth usermodel.AuthUser
	if err := c.ShouldBindJSON(&userAuth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		log.Printf("Error: cannot bind to AuthUser: %s", err.Error())
		c.Abort()
		return
	}
	// Step 2: Validate username and password
	userFound, err := userRepo.GetByUsername(userAuth.Username)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": fmt.Sprintf("Could not find user %q", userAuth.Username)},
		)
		log.Printf("Could not find user %q", userAuth.Username)
		c.Abort()
		return
	}

	if err := authentication.VerifyPassword(userFound.Password, userAuth.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		log.Printf("Error: Invalid password")
		c.Abort()
		return
	}

	// Step 3: Generate JWT/Bearer token
	tkn, err := authentication.NewDefaultTokenString(userFound.Username)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "Could not generate authentication token"},
		)
		log.Printf("Error: Could not generate authentication token: %s", err.Error())
		c.Abort()
		return
	}
	// Step 4: Return bearer token as a response
	c.JSON(http.StatusOK, gin.H{"token": tkn})
	// TODO: Remove, here for testing
	log.Printf("Generated token %s for user '%s'", tkn, userFound.Username)
}

// TODO:
// 1. Make it so each separate file upload is a separate POST request
// so you can give each file a description. Would require the creation of a
// File storage/access api/db
// 2. Handle same-named files
//   - append a (n), where n is the smallest number possible (i.e. file(1), file(2), etc.)
//   - append current date to the file name and strip it when serving it (I like this better)
func postFormUpload(c *gin.Context) {
	owner := c.GetString("username")
	if owner == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "username is empty or not found"})
		log.Println("Error: username is empty or not found")
		c.Abort()
		return
	}

	// debug
	log.Println("Owner: ", owner)

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
