// main.go

package main

import (
    "context"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
    "net/http"
    "os"
    "time"
    "github.com/go-redis/redis/v8"
	"github.com/dgrijalva/jwt-go"
    "time"
)

var (
    ctx             = context.Background()
    userCollection  *mongo.Collection
    organizationCollection *mongo.Collection
    client          *redis.Client
)

var (
    jwtSecret = []byte("your_secret_key_here")
)

func main() {
    router := gin.Default()

    // MongoDB initialization
    mongoURI := os.Getenv("MONGO_URI")
    clientOptions := options.Client().ApplyURI(mongoURI)
    mongoClient, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    defer mongoClient.Disconnect(ctx)

    db := mongoClient.Database("mydb")
    userCollection = db.Collection("users")
    organizationCollection = db.Collection("organizations")

    // Redis initialization
    redisAddr := os.Getenv("REDIS_ADDR")
    client = redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    // Routes
    router.POST("/signup", SignUp)
    router.POST("/signin", SignIn)
    router.POST("/refresh-token", RefreshToken)
    router.POST("/organization", CreateOrganization)
    router.GET("/organization/:organization_id", ReadOrganization)
    router.GET("/organization", ReadAllOrganizations)
    router.PUT("/organization/:organization_id", UpdateOrganization)
    router.DELETE("/organization/:organization_id", DeleteOrganization)
    router.POST("/organization/:organization_id/invite", InviteUserToOrganization)

    router.Run(":8080")
}

func generateAccessToken(user User) (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["user_id"] = user.ID
    claims["exp"] = time.Now().Add(time.Hour * 1).Unix() // Access token expiration time (1 hour)
    accessToken, err := token.SignedString(jwtSecret)
    if err != nil {
        return "", err
    }
    return accessToken, nil
}

func generateRefreshToken(user User) (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["user_id"] = user.ID
    claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix() // Refresh token expiration time (1 week)
    refreshToken, err := token.SignedString(jwtSecret)
    if err != nil {
        return "", err
    }
    return refreshToken, nil
}

func RefreshToken(c *gin.Context) {
    refreshToken := c.PostForm("refresh_token")
    if refreshToken == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
        return
    }

    // Parse and validate the refresh token
    token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
        // Check token signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }

    // Check if the token is valid
    if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // Extract user ID from the token claims
        userID := token.Claims.(jwt.MapClaims)["user_id"].(string)

        // Use the user ID to retrieve user details from your database (e.g., MongoDB)
        // You may want to validate if the user exists and perform any additional checks
        
        // Generate a new access token
        // Assume user is retrieved from the database
        user := User{ID: userID} // Create a dummy user for demonstration
        accessToken, err := generateAccessToken(user)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
            return
        }

        // Respond with the new access token
        c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return
    }
}