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
const RequiredMigration = "16.reward-multi-spend.sql"

// Config holds StarApp service configuration.
type Config struct {
	Auth       map[string]any `koanf:"auth"`
	ConfigDir  string         `koanf:"config_dir"`
	DBDriver   string         `koanf:"db_driver"`
	DBPath     string         `koanf:"db_path"`
	HTTPAddr   string         `koanf:"http_addr"`
	WebUIDir   string         `koanf:"webui_dir"`
	ShowFooter bool           `koanf:"show_footer"`
}

// Load reads config from configDir and optional config.yaml.
func Load(configDir string) (*Config, error) {
	k := koanf.New(".")
	configDir = resolveConfigDir(configDir)
	if err := loadConfigFile(k, configDir); err != nil {
		return nil, err
	}
	if err := applyConfigDefaults(k, configDir); err != nil {
		return nil, err
	}

	var c Config
	if err := k.Unmarshal("", &c); err != nil {
		return nil, err
	}
	if err := finalizeConfig(&c, configDir); err != nil {
		return nil, err
	}
	return &c, nil
}

func resolveConfigDir(configDir string) string {
	if configDir != "" {
		return configDir
	}
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "starapp")
}

func loadConfigFile(k *koanf.Koanf, configDir string) error {
	configPath := filepath.Join(configDir, "config.yaml")
	if !regularFileExists(configPath) {
		return nil
	}
	return k.Load(file.Provider(configPath), yaml.Parser())
}

func regularFileExists(path string) bool {
	f, err := os.Stat(path)
	return err == nil && !f.IsDir()
}

func applyConfigDefaults(k *koanf.Koanf, configDir string) error {
	for key, value := range defaultConfigValues(k, configDir) {
		if err := k.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func defaultConfigValues(k *koanf.Koanf, configDir string) map[string]any {
	defaults := map[string]any{"config_dir": configDir}
	setIfMissing(k, defaults, "db_driver", "sqlite")
	setIfMissing(k, defaults, "db_path", filepath.Join(configDir, "starapp.db"))
	setIfMissing(k, defaults, "show_footer", true)
	return defaults
}

func setIfMissing(k *koanf.Koanf, defaults map[string]any, key string, value any) {
	if k.Get(key) == nil {
		defaults[key] = value
	}
}

func finalizeConfig(c *Config, configDir string) error {
	if err := applyListenAddr(c); err != nil {
		return err
	}
	c.DBPath = resolveUnderConfigDir(c.DBPath, configDir)
	c.WebUIDir = resolveUnderConfigDir(c.WebUIDir, configDir)
	return nil
}

func applyListenAddr(c *Config) error {
	if !useAutoListen(c.HTTPAddr) {
		return nil
	}
	addr, err := listenaddr.AvailableListenAddr()
	if err != nil {
		return err
	}
	c.HTTPAddr = addr
	return nil
}

func resolveUnderConfigDir(path, configDir string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(configDir, path)
}

func useAutoListen(httpAddr string) bool {
	if strings.TrimSpace(os.Getenv("PORT")) != "" {
		return true
	}
	return httpAddr == "" || httpAddr == ":8080"
}
