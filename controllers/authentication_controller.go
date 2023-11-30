package controllers

import (
	"context"
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/dastardlyjockey/restaurant-management-backend/helpers"
	"github.com/dastardlyjockey/restaurant-management-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var validate = validator.New()
var UserCollection = database.Collection(database.Client, "users")

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Println("Failed to hash password")
	}
	return string(hashedPassword)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	check := true
	msg := ""

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err != nil {
		check = false
		msg = fmt.Sprintf("Password mismatch")
	}

	return check, msg
}

func Signup(c *gin.Context) {
	var user models.User

	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validateErr := validate.Struct(user)
	if validateErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": validateErr.Error()})
		return
	}

	// check whether if the email and phone number exist in the database
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking the emails"})
		log.Println(err)
	}

	count, err = UserCollection.CountDocuments(ctx, bson.M{"phone_number": user.Phone})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occurred while checking the phone number"})
		log.Println(err)
	}

	if count > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "The email or phone number already exist"})
		return
	}

	//hash the password
	hashPassword := HashPassword(*user.Password)
	user.Password = &hashPassword

	// Autofill in the extra details
	user.CreatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.UpdatedAt, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		log.Println("Error creating the user timeline: ", err)
	}

	user.ID = primitive.NewObjectID()
	user.UserID = user.ID.Hex()

	//generate the tokens
	token, refreshToken, err := helpers.GenerateAllTokens(*user.Email, *user.FirstName, *user.FirstName, user.UserID)
	if err != nil {
		log.Println("Error generating the tokens: ", err)
	}

	user.Token = &token
	user.RefreshToken = &refreshToken

	//add the user to the database
	result, err := UserCollection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "Failed to register"})
		return
	}

	//response
	c.JSON(http.StatusAccepted, result)
}

func Login(c *gin.Context) {
	var user models.User

	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user from context"})
		return
	}

	// retrieve the user from the database
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var foundUser models.User
	err = UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "The user does not exist"})
		return
	}

	// compare the password
	isPassword, msg := VerifyPassword(*user.Password, *foundUser.Password)
	if !isPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// generate and update the tokens
	token, refreshToken, err := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserID)
	if err != nil {
		log.Println("Error generating token: ", err)
	}

	helpers.UpdateAlTokens(token, refreshToken, foundUser.UserID)

	//response
	c.JSON(http.StatusOK, foundUser)
}
