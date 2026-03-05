package repository
import (
	"database/sql"
	"github.com/google/uuid"
	"MAM/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	user.ID = uuid.New().String() // Generate a UUID for the new user
	query := `INSERT INTO users (id, email, password_hash, name, role)
	VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.Name, user.Role)
	return err
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error){
	user := &models.User{}

	query := `SELECT id, email, password_hash, name, role FROM users WHERE email = ?`

	//QueryRow executes the query and scans the result into the struct
	err := r.db.QueryRow(query,email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
	)
	if err != nil{
		return nil, err //Returns the error if email doesn't found
	}

	return user, nil
}