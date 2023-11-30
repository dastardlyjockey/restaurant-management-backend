package helpers

import (
	"context"
	"fmt"
	"github.com/dastardlyjockey/restaurant-management-backend/controllers"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

type SignedDetails struct {
	FirstName string
	LastName  string
	Uid       string
	Email     string
	jwt.RegisteredClaims
}

var secretKey = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Uid:       uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(30))),
		},
	}

	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(24))),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secretKey))
	if err != nil {
		log.Println(err)
	}

	return token, refreshToken, err
}

func UpdateAlTokens(signedToken string, signedRefreshToken string, userID string) {
	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"update_at", updateAt})

	upsert := true
	filter := bson.M{"user_id": userID}
	opt := options.UpdateOptions{Upsert: &upsert}

	// update the user tokens
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	_, err := controllers.UserCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &opt)
	if err != nil {
		log.Printf("Error updating the token: %v", err)
		return
	}

	return

}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		msg = err.Error()
		return
	}

	//check if the token is valid
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg += err.Error()
		return
	}

	//check expired timestamp
	expUnix := claims.ExpiresAt.Time.Unix()
	if expUnix < time.Now().Local().Unix() {
		msg = fmt.Sprintf("the token has expired")
		return
	}

	return claims, msg
}
