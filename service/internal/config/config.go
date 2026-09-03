package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesread/golure/pkg/listenaddr"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// RequiredMigration is the sql-migrate id this binary expects to be applied.
const RequiredMigration = "14.notification-deliveries.sql"

// Config holds StarApp service configuration.
type Config struct {
	ConfigDir  string         `koanf:"config_dir"`
	DBDriver   string         `koanf:"db_driver"`
	DBPath     string         `koanf:"db_path"`
	HTTPAddr   string         `koanf:"http_addr"`
	WebUIDir   string         `koanf:"webui_dir"`
	ShowFooter bool           `koanf:"show_footer"`
	Auth       map[string]any `koanf:"auth"`
}

// Load reads config from configDir and optional config.yaml.
func Load(configDir string) (*Config, error) {
	k := koanf.New(".")

	if configDir == "" {
		dir, _ := os.UserConfigDir()
		configDir = filepath.Join(dir, "starapp")
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if f, err := os.Stat(configPath); err == nil && !f.IsDir() {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			return nil, err
		}
	}

	k.Set("config_dir", configDir)
	if k.Get("db_driver") == nil {
		k.Set("db_driver", "sqlite")
	}
	if k.Get("db_path") == nil {
		k.Set("db_path", filepath.Join(configDir, "starapp.db"))
	}
	if k.Get("show_footer") == nil {
		k.Set("show_footer", true)
	}

	var c Config
	if err := k.Unmarshal("", &c); err != nil {
		return nil, err
	}

	if useAutoListen(c.HTTPAddr) {
		addr, err := listenaddr.AvailableListenAddr()
		if err != nil {
			return nil, err
		}
		c.HTTPAddr = addr
	}

	if c.DBPath != "" && !filepath.IsAbs(c.DBPath) {
		c.DBPath = filepath.Join(configDir, c.DBPath)
	}

	if c.WebUIDir != "" && !filepath.IsAbs(c.WebUIDir) {
		c.WebUIDir = filepath.Join(configDir, c.WebUIDir)
	}

	return &c, nil
}

func useAutoListen(httpAddr string) bool {
	if strings.TrimSpace(os.Getenv("PORT")) != "" {
		return true
	}
	return httpAddr == "" || httpAddr == ":8080"
}
