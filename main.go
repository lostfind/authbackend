package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	firebaseApp, err := initFirebase()
	if err != nil {
		log.Fatal(err)
	}

	firebaseClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		// 許可したいHTTPメソッドの一覧
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
			"PUT",
			"DELETE",
		},
		// 許可したいHTTPリクエストヘッダの一覧
		AllowHeaders: []string{
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
		},
		AllowOrigins: []string{
			"http://localhost:3000",
		},
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/createUser", func(c *gin.Context) {
		params := (&auth.UserToCreate{}).
			Email(c.PostForm("email")).
			EmailVerified(false).
			Password(c.PostForm("password")).
			Disabled(false)

		u, err := firebaseClient.CreateUser(c.Request.Context(), params)
		if err != nil {
			log.Fatalf("error creating user: %v\n", err)
		}
		log.Printf("Successfully created user: %v\n", u)

		c.JSON(200, gin.H{
			"user": u.UID,
		})
	})

	type JsonRequest struct {
		Email string `json:"email"`
	}

	r.POST("/login", func(c *gin.Context) {
		ctx := c.Request.Context()

		var json JsonRequest
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		email := json.Email

		u, err := firebaseClient.GetUserByEmail(ctx, email)
		if err != nil {
			log.Printf("error getting user by email %s: %v\n", email, err)
		}

		log.Printf("Successfully fetched user data: %v\n", u)

		token, err := firebaseClient.CustomToken(ctx, u.UID)

		if err != nil {
			log.Printf("error minting custom token: %v\n", err)
		}

		c.JSON(200, gin.H{
			"token": token,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func initFirebase() (*firebase.App, error) {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	return app, err
}
