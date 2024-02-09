// auth.go

package main

import (
    "context"
    "encoding/json"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/crypto/bcrypt"
    "log"
    "net/http"
    "time"
)

type User struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type Token struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

func SignUp(c *gin.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Hash the password before storing it
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
        return
    }

    user.Password = string(hashedPassword)

    // Insert the user into MongoDB
    _, err = userCollection.InsertOne(context.Background(), user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}

func SignIn(c *gin.Context) {
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Find the user by email in MongoDB
    err := userCollection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&user)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    // Compare the stored hashed password with the password provided
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(user.Password))
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    // Generate JWT access token and refresh token
    accessToken, err := generateAccessToken(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
        return
    }
    refreshToken, err := generateRefreshToken(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
        return
    }

    // Store refresh token in Redis with expiration time
    err = client.Set(context.Background(), refreshToken, user.Email, time.Hour*24).Err()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
        return
    }

    c.JSON(http.StatusOK, Token{AccessToken: accessToken, RefreshToken: refreshToken})
}
