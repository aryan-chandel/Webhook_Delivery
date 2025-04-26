package routes

import (
	"webhook_delivery/controllers"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.POST("/subscriptions", controllers.AddSubscriber())          //create subscription
	e.GET("/subscriptions/:id", controllers.GetSubscriber())       //  get subscription
	e.POST("/subscriptions/:id", controllers.UpdateSubscriber())   // update subsription
	e.DELETE("/subscriptions/:id", controllers.DeleteSubscriber()) //delete subscription

	e.GET("/subscription/status/:id", controllers.SubscriberStatus()) //get status of recent
	e.GET("/subscription/logs/:id", controllers.SubscriberLog())      //get log
}
