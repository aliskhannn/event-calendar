package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application's configuration structure.
// It encapsulates settings for the server, database, JWT, email, and archiver components.
type Config struct {
	Server   Server   `yaml:"server"`   // Server configuration
	Database Database `yaml:"database"` // Database configuration
	JWT      JWT      `yaml:"jwt"`      // JWT configuration for authentication
	Email    Email    `yaml:"email"`    // Email configuration for SMTP
	Archiver Archiver `yaml:"archiver"` // Archiver configuration for periodic tasks
}

// Server holds configuration for the HTTP server.
type Server struct {
	HTTPPort string `yaml:"httpPort"` // port on which the HTTP server listens
}

// Database holds configuration for connecting to a PostgreSQL database.
type Database struct {
	Host     string // Database host address
	Port     string // Database port
	User     string // Database user
	Password string // Database password
	Name     string // Database name
	SSLMode  string `yaml:"sslmode"` // SSL mode for database connection
}

// JWT holds configuration for JSON Web Token authentication.
type JWT struct {
	Secret string        // Secret key for signing JWTs
	TTL    time.Duration `yaml:"ttl"` // token time-to-live duration
}

// Email holds SMTP configuration for sending emails.
type Email struct {
	SMTPHost string `mapstructure:"smtp_host"` // SMTP server host
	SMTPPort string `mapstructure:"smtp_port"` // SMTP server port
	Username string `mapstructure:"username"`  // SMTP username
	Password string `mapstructure:"password"`  // SMTP password
	From     string `mapstructure:"from"`      // sender email address
}

// Archiver holds configuration for the archiver service.
type Archiver struct {
	Interval time.Duration `yaml:"interval"` // Interval for running the archiver task
}

// DatabaseURL builds a PostgreSQL connection string based on the Database configuration.
// It formats the connection string using the database host, port, user, password, name, and SSL mode.
//
// Returns:
//   - A formatted PostgreSQL connection string.
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// Must loads and parses the application configuration from a YAML file and environment variables.
// It reads the configuration file named "config.yaml" from the "./config" directory and
// overrides specific fields with environment variables. If any error occurs during file reading
// or unmarshaling, it panics with an error message.
//
// Returns:
//   - A pointer to the populated Config struct.
func Must() *Config {
	// Set configuration file details.
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	// Read configuration file.
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("fatal error config file: %s \n", err)
	}

	// Unmarshal configuration into struct.
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Panicf("fatal error config file: %s \n", err)
	}

	// Override database configuration with environment variables.
	cfg.Database.Host = os.Getenv("DB_HOST")
	cfg.Database.Port = os.Getenv("DB_PORT")
	cfg.Database.User = os.Getenv("DB_USER")
	cfg.Database.Password = os.Getenv("DB_PASSWORD")
	cfg.Database.Name = os.Getenv("DB_NAME")

	// Override JWT secret with environment variable.
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	// Override email configuration with environment variables.
	cfg.Email.SMTPHost = os.Getenv("SMTP_HOST")
	cfg.Email.SMTPPort = os.Getenv("SMTP_PORT")
	cfg.Email.Username = os.Getenv("SMTP_USER")
	cfg.Email.Password = os.Getenv("SMTP_PASS")
	cfg.Email.From = os.Getenv("SMTP_FROM")

	return &cfg
}
