package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

var userCollection *mongo.Collection
var secretKey = []byte("secretKey")

func initMongoDB() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Println("Mongo Error")
		log.Fatal("Error in Connecting to Mongodb")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println("Mongo Error")
		log.Fatal(err)
	}
	userCollection = client.Database("clouddrop").Collection("users")
	fmt.Println("Connection Established")
}
func main() {
	fmt.Println("Server Running")
	router := gin.Default()
	initMongoDB()
	router.POST("/register", register)
	router.POST("/login", login)

	router.Run(":8080")
}
func register(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request"})
		return
	}
	if len(user.Username) == 0 || len(user.Password) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or Password Can't be empty"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hashing Error"})
		return
	}
	user.Password = string(hashedPassword)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, _ := userCollection.CountDocuments(ctx, bson.M{"username": user.Username})
	if count > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User Already Exists"})
		return
	}
	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user Created successfully"})

}
func login(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Credentials"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var storedUser User
	err = userCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&storedUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Password"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": storedUser.Username,
	})
	tokenString, err := token.SignedString(secretKey)
	c.JSON(http.StatusOK, gin.H{
		"message": "login Successful",
		"token":   tokenString,
	})
}

/*
func validate(c *gin.Context) {
}
*/
