package web

import (
	"github.com/gin-gonic/gin"

	"secondarymetabolites.org/mibig-api/pkg/models"
)

func (app *application) routes() *gin.Engine {
	api := app.Mux.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/version", app.version)
			v1.GET("/stats", app.stats)
			v1.GET("/repository", app.repository)
			v1.POST("/search", app.search)
			v1.GET("/available/:category/:term", app.available)
			v1.GET("/convert", app.Convert)
			v1.GET("/contributors", app.Contributors)

			v1.POST("/login", app.Login)
			v1.POST("/logout", app.Logout)

			v1.GET("/authtest", app.JWTAuthenticated([]models.Role{}), app.AuthTest)
			v1.POST("/submit", app.JWTAuthenticated([]models.Role{}), app.submit)
			v1.POST("/bgc-registration", app.JWTAuthenticated([]models.Role{}), app.LegacyStoreSubmission)
			v1.POST("/bgc-detail-registration", app.JWTAuthenticated([]models.Role{}), app.LegacyStoreBgcDetailSubmission)
		}
	}

	return app.Mux
}
