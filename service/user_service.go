package service
import (
"errors"
"golang.org/x/crypto/bcrypt"
"MAM/models"
"MAM/repository"
)
type UserService interface {
	Register(req models.RegisterRequest) (*models.User, error)
}
type userService struct {
	repo repository.UserRepository
}
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo}
}
func (s *userService) Register(req models.RegisterRequest) (*models.User, error) {
	// 1. Hash the password securely
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
	return nil, errors.New("failed to hash password")
}
// 2. Create the User model
user := &models.User{
	Email: req.Email,
	PasswordHash: string(hashedPassword),
	Name: req.FullName,
	Role: req.Role,
}
// 3. Save to database using the repository
err = s.repo.CreateUser(user)
if err != nil {
// to return a "Email already exists" message.
return nil, err
}
return user, nil
}