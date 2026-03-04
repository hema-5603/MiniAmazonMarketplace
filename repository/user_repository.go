package repository
import (
	"database/sql"
	"github.com/google/uuid"
	"MAM/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	user.ID = uuid.New().String() // Generate a UUID for the new user
	query := `INSERT INTO users (id, email, password_hash, full_name, role)
	VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.Name, user.Role)
	return err
}