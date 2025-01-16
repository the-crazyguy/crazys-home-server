package environment

import (
	"os"

	"github.com/joho/godotenv"
)

// Loads .env files for the executable.
// See https://github.com/joho/godotenv#precedence--conventions and
// https://github.com/bkeepers/dotenv?tab=readme-ov-file#customizing-rails for
// more info on why multiple are loaded
func Load() {
	env := os.Getenv("SERVER_ENV")
	if env == "" {
		env = "development"
	}
	// env files are loaded from most to least specific
	// if an env file defines a variable, it will NOT be overwritten
	godotenv.Load(".env." + env + ".local")
	if env != "test" {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // .env
}
