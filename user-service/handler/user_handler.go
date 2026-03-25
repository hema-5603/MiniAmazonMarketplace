package handler

import (
	// "encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"user-service/models"
	"user-service/service"

	"github.com/labstack/echo/v4"
	"github.com/golang-jwt/jwt/v5"
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
		slog.Warn("Register Failed: Invalid request payload", slog.String("error",err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request payload",
		})
	}
	// Run a validator here to check email format/password length
	// 2. Call the Service layer
	user, err := h.service.Register(req)

	if err != nil {
		slog.Error("Failed to process registration", slog.String("email",req.Email),slog.String("error",err.Error()))
		//Duplicate email error
		if err.Error() == "This email is already registered"{
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"success":false,
				"message": err.Error(),
			})
			}
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

func (h *UserHandler) Login(c echo.Context)error{
	var req models.LoginRequest

	//1. Bind the JSON Payload
	if err := c.Bind(&req); err!=nil{
		slog.Warn("Login Failed: Invalid request payload", slog.String("error",err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success":"false",
			"message":"Invalid request payload",
		})
	}

	//2. Call the service
	token, err:= h.service.Login(req)
	if err!=nil{
		slog.Warn("Unauthorized login", slog.String("email",req.Email),slog.String("error",err.Error()))
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success":"false",
			"message":err.Error(),
		})
	}

	//3. Return the token to the user
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":"Login Successful",
		"data":map[string]string{
			"token":token,
		},
	})
}


// Get profile
func (h *UserHandler) GetProfile(c echo.Context) error{
	//1. Extract the token placed in the context by the Echo JWT middleware
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)

	//2. Grab the user_id which is embedded during login
	userID := claims["user_id"].(string)

	//3.Fetch the profile
	user, err := h.service.GetProfile(userID)

	if err != nil{
		slog.Error("Failed to retrieve the profile information",slog.String("user_id",userID),slog.String("error",err.Error()))
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success":false,
			"message":err.Error(),
		})
	}

	// 4. Return the data(password_hash will automatically hidden by the model struct)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":true,
		"data":user,
	})
}

func (h *UserHandler) UpdateProfile(c echo.Context)error{
	// 1. Securely extract the ID from the JWT token(IDOR prevention)
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	//2.Bind the incoming JSON
	var req models.UpdateProfileRequest
	if err := c.Bind(&req); err!=nil{
		slog.Warn("Update profile failed: Invalid request payload",slog.String("user_id",userID), slog.String("error",err.Error()))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success":false,
			"message":"Invalid request payload",
		})
	}

	// 3. Pass to the service layer
	updatedUser, err := h.service.UpdateProfile(userID,req)
	if err != nil{
		slog.Error("Failed to process profile update",slog.String("user_id",userID),slog.String("error",err.Error()))
		//return the 409 conflict for duplicate email error
		if err.Error() == "This Email is already in use by another account"{
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"success":false,
				"message":err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success":false,
			"message":err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":"Profile updated successfully",
		"data":updatedUser,
	})
}
