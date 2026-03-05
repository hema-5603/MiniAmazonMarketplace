package service
import (
"errors"
"os"
"time"
"golang.org/x/crypto/bcrypt"
"github.com/golang-jwt/jwt/v5"
"MAM/models"
"MAM/repository"
)

type UserService interface {
	Register(req models.RegisterRequest) (*models.User, error)
	Login(req models.LoginRequest) (string,error)
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

	//3.Generate the JWT token
	//Fetch the secret from the environment
	secret := os.Getenv("jwtSecKey")

	if secret == ""{
		//Fallback or error if the key is missing in .env
		return "", errors.New("Internal server error: missing secret key")
	}

	jwtSecretKey := []byte(secret)
	//Embedding the user ID and Role directly into the token payload(claims)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":user.Role,
		"exp": time.Now().Add(time.Hour*24).Unix(), //Token will expire in 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	tokenString,err := token.SignedString(jwtSecretKey)

	if err != nil{
		return "", errors.New("Failed to generate token")
	}
	return tokenString,nil
}