package main

import "github.com/gin-gonic/gin"

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

			v1.POST("/submit", app.submit)
			v1.POST("/bgc-registration", app.LegacyStoreSubmission)
			v1.POST("/bgc-detail-registration", app.LegacyStoreBgcDetailSubmission)
		}
	}

	return app.Mux
}
