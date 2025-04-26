package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"webhook_delivery/controllers"
	"webhook_delivery/service"

	"github.com/labstack/echo/v4"
)

func VerifySignature() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			subscriptionID := c.Param("sub_id")

			// 1. Get subscription from cache or DB
			sub, err := controllers.GetSubscriptionByID(subscriptionID)
			if err != nil {
				return c.JSON(http.StatusNotFound, echo.Map{"error": "subscription not found"})
			}

			// 2. If no secret, skip check
			if sub.Secret == "" {
				return next(c)
			}

			// 3. Read raw body
			bodyBytes, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "could not read body"})
			}
			c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset body for next handler

			// 4. Get signature header from request
			receivedSig := c.Request().Header.Get("X-Signature")
			if receivedSig == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing signature header"})
			}

			// 5. Compute HMAC
			computedSig := service.ComputeHMAC(bodyBytes, sub.Secret)

			//6.verify signature
			if strings.ToLower(computedSig) != strings.ToLower(receivedSig) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
			}

			return next(c)
		}
	}

}
