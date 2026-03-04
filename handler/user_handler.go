package handler
import (
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
	"MAM/models"
	"MAM/service"
)
type UserHandler struct {
	service service.UserService
}
func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}
func (h *UserHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	// 1. Bind JSON payload to the struct
	if err := c.Bind(&req); err != nil {
	return c.JSON(http.StatusBadRequest, map[string]interface{}{
	"success": false,
	"message": "Invalid request payload",
	})
}
// Run a validator here to check email format/password length
// 2. Call the Service layer
user, err := h.service.Register(req)
if err != nil {
	fmt.Println("DB ERROR:",err)
	return c.JSON(http.StatusInternalServerError, map[string]interface{}{
	"success": false,
	"message": "Failed to register user",
	})
}
// 3. Return 201 Created
return c.JSON(http.StatusCreated, map[string]interface{}{
	"message": "User registered successfully",
	"data": map[string]string{
	"id": user.ID,
	},
})
}
