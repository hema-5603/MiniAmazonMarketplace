package service
import (
"errors"
"log/slog"
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
	UpdateProfile(userID string, req models.UpdateProfileRequest) (*models.User, error)
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
	// Check the email isn't already taken
	existingUser, _ := s.repo.GetUserByEmail(req.Email)
	if existingUser != nil{
		slog.Warn("Registation failed: Email already exists",slog.String("email",req.Email))
		return nil, errors.New("This email is already registered")
	}
	// 1. Hash the password securely
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Warn("Failed to hash password",slog.String("error",err.Error()))
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
		slog.Error("Database error during User creation",slog.String("error",err.Error()),slog.String("email",req.Email))
		// to return a "Email already exists" message.
		return nil, err
	}
	//Success log!
	slog.Info("New user successfully registered",slog.String("user_id",user.ID), slog.String("role", user.Role))
	return user, nil
}

func (s *userService) Login(req models.LoginRequest) (string, error){
	// 1. Check if the user exists
	user, err := s.repo.GetUserByEmail(req.Email)
	if err!= nil{
		//Instead of directly telling the user not found, use generic message so hackers can't guess valid emails
		slog.Warn("Login failed: User not found",slog.String("email",req.Email))
		return "",errors.New("Invalid email or password")
	}
	// 2. Compare the provided password with the hash password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err!= nil{
		slog.Warn("Login failed: password",slog.String("email",req.Email))
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
		slog.Error("Failed to generate JWT", slog.String("user_id",user.ID), slog.String("error",err.Error()))
		return "", errors.New("Failed to generate token")
	}
	slog.Info("User logged in successfully", slog.String("user_id",user.ID))
	return tokenString,nil
}

// Get profile Information

func (s *userService) GetProfile(userID string) (*models.User, error){
	user, err := s.repo.GetUserByID(userID)

	if err!=nil{
		slog.Warn("Profile fetch failed: User not found", slog.String("user_id",userID))
		return nil, errors.New("User not found")
	}
	slog.Info("Profile fetched successfully", slog.String("user_id", userID))
	return user, nil
}

func (s *userService) UpdateProfile(userID string, req models.UpdateProfileRequest) (*models.User,error){
	//1. Fetch the user's current data from the DB
	user, err := s.repo.GetUserByID(userID)
	if err != nil{
		slog.Warn("Update failed: User not found", slog.String("user_id",userID))
		return nil, errors.New("User not found")
	}
	//2. Check for Email duplication
	if req.Email != "" && req.Email != user.Email{
		existingUser, _ := s.repo.GetUserByEmail(req.Email)
		if existingUser != nil{
			slog.Warn("Update failed: Email already in use", slog.String("user_id",userID), slog.String("attempted_email",req.Email))
			return nil, errors.New("This Email is already in use by another account")
		}
		user.Email = req.Email
	}

	// 3. Update name
	if req.Name != ""{
		user.Name = req.Name
	}

	// 4. Update securely hash new password
	if req.Password != ""{
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password),bcrypt.DefaultCost)
		
		if err != nil{
			slog.Error("Failed to hash new password", slog.String("user_id",userID), slog.String("error",err.Error()))
			return nil, errors.New("Failed to secure new password")
		}
		user.PasswordHash = string(hashedPassword)
	}

	//5. Save the updated user back to the database
	err = s.repo.UpdateUser(user)
	if err != nil{
		slog.Error("Database error during profile update",slog.String("user_id",userID), slog.String("error",err.Error()) )
		return nil, errors.New("Failed to update profile")
	}
	slog.Info("User profile updated successfully", slog.String("user_id",userID))
	return user, nil
}