package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/jus1d/kypidbot/internal/config/messages"
)

const (
	EnvLocal       = "local"
	EnvDevelopment = "dev"
	EnvProduction  = "prod"
)

type Config struct {
	Env           string        `yaml:"env" env-required:"true"`
	Bot           Bot           `yaml:"bot" env-required:"true"`
	Ollama        Ollama        `yaml:"ollama" env-required:"true"`
	Postgres      Postgres      `yaml:"postgres" env-required:"true"`
	S3            S3            `yaml:"s3" env-required:"true"`
	Notifications Notifications `yaml:"notifications"`
}

type Bot struct {
	Token        string `yaml:"token" env-required:"true"`
	MessagesPath string `yaml:"messages_path" env-required:"true"`
}

type Notifications struct {
	PollInterval           time.Duration `yaml:"poll_interval" env-default:"5s"`
	DateUpcomingIn         time.Duration `yaml:"date_upcoming_in" env-default:"1h"`
	RegistrationReminderIn time.Duration `yaml:"registration_reminder_in" env-default:"24h"`
	InviteReminderIn       time.Duration `yaml:"invite_reminder_in" env-default:"10m"`
}

type Ollama struct {
	Host      string `yaml:"host" env-required:"true"`
	Port      string `yaml:"port" env-required:"true"`
	Model     string `yaml:"model" env-required:"true"`
	MaxLength int    `yaml:"max_length" env-default:"512"`
}

type Postgres struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Name     string `yaml:"name" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	ModeSSL  string `yaml:"sslmode" env-required:"true"`
}

type S3 struct {
	Host            string `yaml:"host" env-required:"true"`
	Port            string `yaml:"port" env-required:"true"`
	AccessKeyID     string `yaml:"access_key_id" env-required:"true"`
	SecretAccessKey string `yaml:"secret_access_key" env-required:"true"`
	Bucket          string `yaml:"bucket" env-required:"true"`
	Region          string `yaml:"region" env-default:"us-east-1"`
	UseSSL          bool   `yaml:"use_ssl" env-default:"false"`
}

// MustLoad loads config to a new Config instance and return it
func MustLoad() *Config {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		panic("missed CONFIG_PATH environment variable")
	}

	var err error
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var config Config

	if err = cleanenv.ReadConfig(configPath, &config); err != nil {
		panic("cannot read config: " + err.Error())
	}

	if err = cleanenv.ReadConfig(config.Bot.MessagesPath, &messages.M); err != nil {
		panic("cannot read messages: " + err.Error())
	}

	return &config
}
