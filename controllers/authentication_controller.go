package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func HashPassword(password string) string {}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {}

func Signup(c *gin.Context) {}

func Login(c *gin.Context) {}
