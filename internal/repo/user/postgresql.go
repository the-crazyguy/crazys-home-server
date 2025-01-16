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
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	query := `SELECT trusted_user_id FROM user_trusts WHERE user_id = $1`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get trusted users: %w", err)
	}
	defer rows.Close()

	var trustedIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("Failed to scan trustedUserID: %w", err)
		}
		trustedIDs = append(trustedIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Failed to iterate over returned trusted_user_id rows: %w", err)
	}

	return trustedIDs, nil
}

func (r *PostgresUserRepository) GetUserIDsTrusting(userID string) ([]string, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	query := `SELECT user_id FROM user_trusts WHERE trusted_user_id = $1`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get trusting users: %w", err)
	}
	defer rows.Close()

	var trustingUserIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("Failed to scan trusting user: %w", err)
		}
		trustingUserIDs = append(trustingUserIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Failed to iterate over returned trusted_user_id rows: %w", err)
	}

	return trustingUserIDs, nil
}

func (r *PostgresUserRepository) CreateTrust(userID, trustedUserID string) error {
	if userID == "" {
		return errors.New("userID cannot be empty")
	}
	if trustedUserID == "" {
		return errors.New("trustedUserID cannot be empty")
	}

	// NOTE: Add `ON CONFLICT DO NOTHING`?
	query := `INSERT INTO user_trusts (user_id, trusted_user_id) VALUES ($1, $2)`

	_, err := r.db.Exec(context.Background(), query, userID, trustedUserID)
	if err != nil {
		return fmt.Errorf("Erorr creating trust: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) DeleteTrust(userID, otherID string) error {
	if userID == "" {
		return errors.New("userID cannot be empty")
	}
	if otherID == "" {
		return errors.New("otherID cannot be empty")
	}

	query := `DELETE FROM user_trusts WHERE user_id = $1 AND trusted_user_id = $2`

	cmdTag, err := r.db.Exec(context.Background(), query, userID, otherID)
	if err != nil {
		return fmt.Errorf("Failed to delete trust: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("Failed to delete trust, no rows affected")
	}

	return nil
}

func (r *PostgresUserRepository) Trusts(userID, targetID string) (bool, error) {
	if userID == "" {
		return false, errors.New("Empty userID")
	}
	if targetID == "" {
		return false, errors.New("Empty targetID")
	}

	query := `SELECT EXISTS(SELECT 1 FROM user_trusts WHERE user_id = $1 AND trusted_user_id = $2)`

	var trustExists bool
	if err := r.db.QueryRow(context.Background(), query, userID, targetID).Scan(&trustExists); err != nil {
		return false, fmt.Errorf("Failed executing query: %w", err)
	}

	return trustExists, nil
}
