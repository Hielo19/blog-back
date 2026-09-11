package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

const DefaultPath = "configs/config.yaml"

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

func Load(path string) (Config, error) {
	configuration := Config{
		Server: ServerConfig{Port: 8080},
		Database: DatabaseConfig{
			Host:    "127.0.0.1",
			Port:    5432,
			SSLMode: "disable",
		},
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("读取配置文件 %q: %w", path, err)
	}
	if err := yaml.Unmarshal(content, &configuration); err != nil {
		return Config{}, fmt.Errorf("解析配置文件 %q: %w", path, err)
	}
	if err := configuration.validate(); err != nil {
		return Config{}, fmt.Errorf("配置文件 %q: %w", path, err)
	}

	return configuration, nil
}

func (c Config) Address() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

func (c Config) DatabaseURL() string {
	userInfo := url.User(c.Database.User)
	if c.Database.Password != "" {
		userInfo = url.UserPassword(c.Database.User, c.Database.Password)
	}

	query := url.Values{}
	query.Set("sslmode", c.Database.SSLMode)

	return (&url.URL{
		Scheme:   "postgres",
		User:     userInfo,
		Host:     net.JoinHostPort(c.Database.Host, strconv.Itoa(c.Database.Port)),
		Path:     "/" + c.Database.Name,
		RawQuery: query.Encode(),
	}).String()
}

func (c Config) validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("server.port 必须是 1 到 65535 之间的整数")
	}
	if strings.TrimSpace(c.Database.Host) == "" {
		return errors.New("database.host 不能为空")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return errors.New("database.port 必须是 1 到 65535 之间的整数")
	}
	if strings.TrimSpace(c.Database.User) == "" {
		return errors.New("database.user 不能为空")
	}
	if strings.TrimSpace(c.Database.Name) == "" {
		return errors.New("database.name 不能为空")
	}

	validSSLModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[c.Database.SSLMode] {
		return errors.New("database.ssl_mode 配置值无效")
	}

	return nil
}
