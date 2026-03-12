package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"user-service/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1. Create a Mock Service
type MockUserService struct{
	mock.Mock
}

func (m *MockUserService) Register(req models.RegisterRequest) (*models.User, error){
	args := m.Called(req)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(req models.LoginRequest) (string, error) {
	args := m.Called(req)
	return args.String(0),args.Error(1)
}
func (m *MockUserService) GetProfile(userID string) (*models.User,error) {
	args := m.Called(userID)
	return args.Get(0).(*models.User),args.Error(1)
}
func (m *MockUserService) UpdateProfile(userID string, req models.UpdateProfileRequest)(*models.User, error){
	args := m.Called(userID, req)
	return args.Get(0).(*models.User),args.Error(1)
}

//Helper function to inject a fake JWT token into the echo context
func setJWTContext(c echo.Context, userID string){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id":userID})
	c.Set("user",token)
}

//2.Tests

func TestRegisterHandler_Success(t *testing.T){
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	mockUser := &models.User{ID:"123",Email: "gojo@gmail.com"}
	mockService.On("Register",mock.Anything).Return(mockUser,nil)

	reqBody := `{"email":"gojo@gmail.com","password":"bakabaka29","name":"Gojo Sensei","role":"SELLER"}`
	req := httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType,echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Register(c)

	assert.NoError(t,err)
	assert.Equal(t,http.StatusCreated,rec.Code)
	assert.Contains(t,rec.Body.String(),"User registered successfully")
}
func TestLoginHandler_Success(t *testing.T){
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	mockService.On("Login",mock.Anything).Return("fake-jwt-token",nil)

	reqBody := `{"email":"gojo@gmail.com","password":"bakabaka29"}`
	req := httptest.NewRequest(http.MethodPost,"/api/v1/auth/login",bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType,echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Login(c)

	assert.NoError(t,err)
	assert.Equal(t,http.StatusOK,rec.Code)
	assert.Contains(t,rec.Body.String(),"fake-jwt-token")
}

func TestGetProfileHandler_Success(t *testing.T){
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	mockUser := &models.User{ID:"123",Email: "gojo@gmail.com",Name: "Gojo Sensei"}
	mockService.On("GetProfile",mock.Anything).Return(mockUser,nil)

	req := httptest.NewRequest(http.MethodGet,"/api/v1/users/profile",nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//Inject the mock JWT token before calling the handler
	setJWTContext(c,"123")
	err := handler.GetProfile(c)

	assert.NoError(t,err)
	assert.Equal(t,http.StatusOK,rec.Code)
	assert.Contains(t,rec.Body.String(),"Gojo Sensei")
}

func TestUpdateProfileHandler_Success(t *testing.T){
	//1.Setup echo and Mock
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)

	//2.Define the expected mock behaviour
	expectedUser := &models.User{ID:"123",Email: "gojo@gmail.com",Name: "Satoru Gojo"}
	reqPayLoad := models.UpdateProfileRequest{Name: "Satoru Gojo"}
	mockService.On("UpdateProfile","123",reqPayLoad).Return(expectedUser,nil)

	//3. Create a fake HTTP PUT request with a JSON body
	reqBody := `{"name":"Satoru Gojo"}`
	req := httptest.NewRequest(http.MethodPut,"/api/v1/users/profile",bytes.NewBufferString(reqBody))
	req.Header.Set(echo.HeaderContentType,echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	//4. Inject the JWT token into the context(It's critical for protected routes)
	setJWTContext(c,"123")

	//5.Execute the handler
	err := handler.UpdateProfile(c)

	//6. Assertions
	assert.NoError(t,err)
	assert.Equal(t,http.StatusOK,rec.Code)
	assert.Contains(t,rec.Body.String(),"Profile updated successfully")
	assert.Contains(t,rec.Body.String(), "Satoru Gojo")

	mockService.AssertExpectations(t)
}