package main

import (
	"webhook_delivery/controllers"
	"webhook_delivery/middleware"
	"webhook_delivery/routes"
	"webhook_delivery/service"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	//start background redis
	go service.StartWorker()
	//start log cleaner
	go service.StartLogRententionWorker()
	routes.RegisterRoutes(e)
	e.POST("/ingest/:sub_id", controllers.NewDelivery(), middleware.VerifySignature())
	//demo endpoint for engine testing
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "status working"})
	})
	e.Logger.Fatal(e.Start(":8000"))
}
