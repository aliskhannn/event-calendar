package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	JWT      JWT      `yaml:"jwt"`
	Email    Email    `yaml:"email"`
	Archiver Archiver `yaml:"archiver"`
}

type Server struct {
	HTTPPort string `yaml:"httpPort"`
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string `yaml:"sslmode"`
}

type JWT struct {
	Secret string
	TTL    time.Duration `yaml:"ttl"`
}

// Email holds SMTP configuration for sending emails.
type Email struct {
	SMTPHost string `mapstructure:"smtp_host"`
	SMTPPort string `mapstructure:"smtp_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type Archiver struct {
	Interval time.Duration `yaml:"interval"`
}

// DatabaseURL builds a PostgreSQL connection string
// based on the Database configuration.
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

func Must() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("fatal error config file: %s \n", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Panicf("fatal error config file: %s \n", err)
	}

	cfg.Database.Host = os.Getenv("DB_HOST")
	cfg.Database.Port = os.Getenv("DB_PORT")
	cfg.Database.User = os.Getenv("DB_USER")
	cfg.Database.Password = os.Getenv("DB_PASSWORD")
	cfg.Database.Name = os.Getenv("DB_NAME")

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	cfg.Email.SMTPHost = os.Getenv("SMTP_HOST")
	cfg.Email.SMTPPort = os.Getenv("SMTP_PORT")
	cfg.Email.Username = os.Getenv("SMTP_USERNAME")
	cfg.Email.Password = os.Getenv("SMTP_PASSWORD")
	cfg.Email.From = os.Getenv("EMAIL_FROM")

	return &cfg
}
