package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"news-collector/constants"
	"os"
)

// Config exported
type Config struct {
	Application       AppConfig         `json:"app"`
	Memphis           MemphisConfig     `json:"memphis"`
	TheGuardianConfig TheGuardianConfig `json:"theGuardianConfig"`
	TheNewsApiConfig  TheNewsApiConfig  `json:"theNewsApiConfig"`
	CollectorSettings CollectorSettings `json:"collectorSettings"`
}

// AppConfig exported
type AppConfig struct {
	Name    string `json:"name"`
	Env     string `env:"DEPLOY_ENV" env-required:"required"`
	Version string `json:"version"`
}

type TheGuardianConfig struct {
	BaseUrl     string `json:"baseUrl"`
	GetNewsPath string `json:"getNewsPath"`
	HttpMethod  string `json:"httpMethod"`
	ApiKey      string `json:"apiKey" env:"GUARDIAN_API_KEY" env-required:"required"`
}

type TheNewsApiConfig struct {
	BaseUrl     string `json:"baseUrl"`
	GetNewsPath string `json:"getNewsPath"`
	HttpMethod  string `json:"httpMethod"`
}

type CollectorSettings struct {
	SourceOfNews   string `json:"sourceOfNews" env:"SOURCE_OF_NEWS" env-required:"required"`
	ScrapeInterval string `json:"scrapeInterval" env:"SCRAPE_INTERVAL" env-required:"required"`
}

// MemhpisConfig exported
type MemphisConfig struct {
	Host        string `json:"host" env:"MEMPHIS_HOST" env-required:"required"`
	Username    string `json:"username" env:"MEMPHIS_USERNAME" env-required:"required"`
	Token       string `json:"token" env:"MEMPHIS_TOKEN" env-required:"required"`
	Port        int    `json:"port" env:"MEMPHIS_PORT" envDefault:"6666"`
	NewsStation string `json:"newsStation" env:"STATION_NEWS_ID" env-required:"required"`
}

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	// Local development purpose
	if env := os.Getenv(constants.DEPLOY_ENV_KEY); env == "" || env == "LOCAL" {
		errLoad := godotenv.Load(".env.local")

		if errLoad != nil {
			log.Fatalln(errLoad)
		}
	}

	cfg := &Config{}

	pathConfigFile := os.Getenv(constants.APP_CONFIG_PATH_ENV_KEY)

	if pathConfigFile == "" {
		return nil, fmt.Errorf("config error: please provide environment variable: %s",
			constants.APP_CONFIG_PATH_ENV_KEY)
	}

	if _, err := os.Stat(pathConfigFile); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err := cleanenv.ReadConfig(pathConfigFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = checkEnvironExistence()
	if err != nil {
		return nil, err
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetConfig(ctx context.Context) (*Config, error) {
	cfgExist := ctx.Value(constants.CONFIG_CTX_KEY)
	if cfgExist == nil {
		return nil, fmt.Errorf("config errror: unable to find config from context by key: %s \n", constants.CONFIG_CTX_KEY)
	}

	if cfg, ok := cfgExist.(*Config); ok {
		return cfg, nil
	}

	return nil, fmt.Errorf(
		"config errror: failed to convert element's value: %s into desired type: %s \n",
		constants.CONFIG_CTX_KEY, "*Config")
}

func PutConfig(ctx context.Context, cfg *Config) context.Context {
	newCtx := context.WithValue(ctx, constants.CONFIG_CTX_KEY, cfg)
	return newCtx
}

func checkEnvironExistence() error {
	for _, envKey := range constants.ALL_CONFIG_ENV_KEYS {
		if envVal := os.Getenv(envKey); envVal == "" {
			return fmt.Errorf(
				"config error: missing environment variable: %s",
				envKey)
		}
	}
	return nil
}
