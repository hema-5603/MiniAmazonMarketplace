package service

import (
	"testing"

	"user-service/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// 1. Create a Mock Repository that satisfies your UserRepository interface
type MockUserRepository struct{
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error{
	args := m.Called(user)
	return args.Error(0)
}

//Include the other interface methods even if we don't use them, so that go doesn't complain
func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args:=m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User),args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) GetUserByID(id string) (*models.User, error){
	args:=m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User),args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) UpdateUser(user *models.User) error{
	args := m.Called(user)
	return args.Error(0)
}

//2. Tests 
// 2.1 Register function
func TestRegister_Success(t *testing.T){
	//Setup the mock repo and the service
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo,"test_secret")

	//Create a fake incoming request
	req := models.RegisterRequest{
		Email: "gojo@gmail.com",
		Password: "bakabaka29",
		Name: "Gojo Sensei",
		Role:"SELLER",
	}

	//Tell the mock: "When CreateUser is called with ANY user model, return nil (no error)"
	mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	//Execute the actual service function
	user, err := userService.Register(req)

	//Assertions (Did it behave as expected?)
	assert.NoError(t,err) //No error should get
	assert.NotNil(t,user) //The user object shouldn't be empty
	assert.Equal(t,req.Email,user.Email) //The email should match
	assert.NotEmpty(t,user.PasswordHash) //The password should be hashed

	//Verify that the mock was actually called
	mockRepo.AssertExpectations(t)
}

//2.2 Login function
func TestLogin_Success(t *testing.T){
	//Setup the mock repo and the service
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo,"test_secret")

	//Hash the password in the mock to bcrypt.Compare succeeds
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("bakabaka29"),bcrypt.DefaultCost)
	mockUser := &models.User{ID:"123",Email: "gojo@gmail.com",PasswordHash: string(hashedPassword),Role:"SELLER"}
	//Create a fake incoming request
	req := models.LoginRequest{
		Email: "gojo@gmail.com",
		Password: "bakabaka29",
	}

	//Tell the mock: "When CreateUser is called with ANY user model, return nil (no error)"
	mockRepo.On("GetUserByEmail",req.Email).Return(mockUser,nil)

	//Execute the actual service function
	token, err := userService.Login(req)

	//Assertions (Did it behave as expected?)
	assert.NoError(t,err) //No error should get
	assert.NotEmpty(t,token) //The password should be hashed

	//Verify that the mock was actually called
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_Success(t *testing.T){
	//Setup the mock repo and the service
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo,"test_secret")
	mockUser := &models.User{ID:"123",Email: "gojo@gmail.com",Name: "Gojo Sensei"}
	mockRepo.On("GetUserByID","123").Return(mockUser, nil)

	user, err := userService.GetProfile("123")

	assert.NoError(t,err)
	assert.Equal(t,"Gojo Sensei",user.Name)
	mockRepo.AssertExpectations(t)

}

func TestUpdateProfile_Success(t *testing.T){
	//Setup the mock repo and the service
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo,"test_secret")

	existingUser := &models.User{ID:"123",Email: "gojo@gmail.com",Name: "Gojo Sensei"}
	req := models.UpdateProfileRequest{Name: "Satoru Gojo"}
	
	mockRepo.On("GetUserByID","123").Return(existingUser, nil)
	mockRepo.On("UpdateUser", mock.AnythingOfType("*models.User")).Return(nil)

	user, err := userService.UpdateProfile("123",req)

	assert.NoError(t,err)
	assert.Equal(t,"Satoru Gojo",user.Name)
	mockRepo.AssertExpectations(t)

}
