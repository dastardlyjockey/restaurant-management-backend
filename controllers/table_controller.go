package controllers

import (
	"github.com/dastardlyjockey/restaurant-management-backend/database"
	"github.com/gin-gonic/gin"
)

var tableCollection = database.Collection(database.Client, "table")

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetTableById() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
