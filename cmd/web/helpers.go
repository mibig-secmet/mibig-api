package main

import (
	"github.com/gin-gonic/gin"
	zap "go.uber.org/zap"
	"net/http"
)

func (app *application) clientError(c *gin.Context, status int) {
	c.JSON(status, gin.H{"message": http.StatusText(status)})
}

func (app *application) serverError(c *gin.Context, err error) {
	app.logger.Errorw("server error", zap.Error(err))
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": http.StatusText(http.StatusInternalServerError)})
}

func (app *application) notFound(c *gin.Context) {
	app.clientError(c, http.StatusNotFound)
}
