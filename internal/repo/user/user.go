package user

import usermodel "crzy-server/internal/models/user"

type UserRepository interface {
	GetByID(id string) (*usermodel.User, error)
	GetByUsername(username string) (*usermodel.User, error)
	Create(user *usermodel.User) error
	Update(user *usermodel.User) error
	// TODO: Add a soft-delete option
	Delete(id string) error

	// Trust logic: doesn't make much sense to separate into its own repo as of now

	// Retrieves the IDs of the users that the provided userID trusts
	GetTrustedUserIDs(userID string) ([]string, error)
	// Creates a trust. userID trusts trustedUserID
	CreateTrust(userID, trustedUserID string) error
	// Deletes a trust relationship. userID no longer trusts otherID
	DeleteTrust(userID, otherID string) error
}
