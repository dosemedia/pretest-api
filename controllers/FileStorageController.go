package controllers

import (
	"log"
	"net/http"

	"aaronblondeau.com/hasura-base-go/prisma/db"
	"github.com/labstack/echo/v4"
)

type FileStorageController struct {
	client *db.PrismaClient
}

func NewFileStorageController(e *echo.Echo) *FileStorageController {
	controller := FileStorageController{
		client: db.NewClient(),
	}
	return &controller
}

func (controller *FileStorageController) Run(e *echo.Echo) error {

	if err := controller.client.Prisma.Connect(); err != nil {
		return err
	}

	e.POST("/files/user-avatar", func(c echo.Context) error {
		// File field is named "avatar"

		return c.JSON(http.StatusOK, map[string]interface{}{
			"todo": true,
		})
	})

	e.GET("/files/user-avatar/:userId/:fileId", func(c echo.Context) error {
		userId := c.Param("userId")
		fileId := c.Param("fileId")

		// TODO
		return c.JSON(http.StatusOK, map[string]interface{}{
			"todo":   true,
			"userId": userId,
			"fileId": fileId,
		})
	})

	return nil
}

func (controller *FileStorageController) Shutdown() {
	log.Print("Shutting down file storage controller...")
	if controller.client != nil {
		controller.client.Disconnect()
	}
}
