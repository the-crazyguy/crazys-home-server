package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	usermodel "crzy-server/internal/models/user"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

// TODO: Function to close the db connection?
func NewPostgresUserRepository(config *pgxpool.Config) (*PostgresUserRepository, error) {
	// Example for a config:
	// exampleConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	// if err != nil {
	// 	log.Println("Unable to parse DATABASE_URL:", err)
	// }

	if config == nil {
		return nil, errors.New("Missing pgx config")
	}

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return &PostgresUserRepository{db: db}, nil
}

func (r *PostgresUserRepository) GetByID(id string) (*usermodel.User, error) {
	// TODO: Validate UUID?
	query := `SELECT id, username, password, created_at FROM users WHERE id = $1`

	row := r.db.QueryRow(context.Background(), query, id)

	var user usermodel.User
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetByUsername(username string) (*usermodel.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	query := `SELECT id, username, password, created_at FROM users WHERE username = $1`

	row := r.db.QueryRow(context.Background(), query, username)

	var user usermodel.User
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresUserRepository) Create(user *usermodel.User) error {
	if err := usermodel.ValidateUser(user); err != nil {
		return err
	}

	query := `INSERT INTO users (id, username, password, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(context.Background(), query,
		user.ID,
		user.Username,
		user.Password,
		user.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("Erorr creating user: %w", err)
	}

	return nil
}

// Updates the fields of the provided user.
// `user`'s ID must already exist. CreatedAt will not be modified, even if it is different
func (r *PostgresUserRepository) Update(user *usermodel.User) error {
	query := `UPDATE users SET username = $1, password = $2 WHERE id = $3`

	cmdTag, err := r.db.Exec(context.Background(), query, user.Username, user.Password, user.ID)
	if err != nil {
		return fmt.Errorf("Failed to update user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("User with id '%s' not found", user.ID)
	}

	return nil
}

func (r *PostgresUserRepository) Delete(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	query := `DELETE FROM users WHERE id = $1`

	cmdTag, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("Failed to delete user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user with id '%s' not found", id)
	}

	return nil
}

func (r *PostgresUserRepository) GetTrustedUserIDs(userID string) ([]string, error) {
	return nil, errors.New("Not implemented")
}

func (r *PostgresUserRepository) CreateTrust(userID, trustedUserID string) error {
	return errors.New("Not implemented")
}

func (r *PostgresUserRepository) DeleteTrust(userID, otherID string) error {
	return errors.New("Not implemented")
}
