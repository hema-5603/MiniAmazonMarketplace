package service
import (
"errors"
"time"
"golang.org/x/crypto/bcrypt"
"github.com/golang-jwt/jwt/v5"
"user-service/models"
"user-service/repository"
)

type UserService interface {
	Register(req models.RegisterRequest) (*models.User, error)
	Login(req models.LoginRequest) (string,error)
	GetProfile(userID string) (*models.User, error)
}
type userService struct {
	repo repository.UserRepository
	jwtSecret string
}
func NewUserService(repo repository.UserRepository, secret string) UserService {
	return &userService{
		repo : repo,
		jwtSecret: secret,
	}
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
		Name: req.Name,
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

func (s *userService) Login(req models.LoginRequest) (string, error){
	// 1. Check if the user exists
	user, err := s.repo.GetUserByEmail(req.Email)
	if err!= nil{
		//Instead of directly telling the user not found, use generic message so hackers can't guess valid emails
		return "",errors.New("Invalid email or password")
	}
	// 2. Compare the provided password with the hash password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err!= nil{
		return "", errors.New("Invalid email or password")
	}

	//Embedding the user ID and Role directly into the token payload(claims)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":user.Role,
		"exp": time.Now().Add(time.Hour*24).Unix(), //Token will expire in 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	tokenString,err := token.SignedString([]byte(s.jwtSecret))

	if err != nil{
		return "", errors.New("Failed to generate token")
	}
	return tokenString,nil
}

// Get profile Information

func (s *userService) GetProfile(userID string) (*models.User, error){
	user, err := s.repo.GetUserByID(userID)

	if err!=nil{
		return nil, errors.New("User not found")
	}
	return user, nil
}