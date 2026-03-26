package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func RoleBasedAccess(allowedRoles ...string) echo.MiddlewareFunc{
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. Get the parsed JWT from the echo context
			userToken, ok := c.Get("user").(*jwt.Token)
			if !ok{
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message" : "Missing or invalid token", 
				})
			} 
			// 2. Extract the claims
			claims, ok := userToken.Claims.(jwt.MapClaims)
			if !ok{
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message" : "Invalid token claims",
				})
			}
			// 3. Get the user role
			userRole, ok := claims["role"].(string)
			if !ok{
				return c.JSON(http.StatusForbidden, map[string]string{
					"message" : "Role not found in token",
				})
			}

			// 4. Check if their role matches any of the allowed roles
			for _, allowedRole := range allowedRoles{
				if userRole == allowedRole{
					// Success : Pass them to the actual handler (Checkout/cancel)
					return next(c)
				}
			}

			// Failure: Don't allow
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"success" : false,
				"message" : "Forbidden: You do not have the required role to access this service",
			})
		}
	}
}