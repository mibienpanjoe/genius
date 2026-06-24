package workspace

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the persisted genius settings (docs/05 §config). Stored at
// $XDG_CONFIG_HOME/genius/config.toml (else ~/.config/genius/config.toml).
type Config struct {
	StudyRoot     string `toml:"study_root"`
	DefaultEngine string `toml:"default_engine"`
	Model         string `toml:"model"`
}

// configPath returns the config file location, honoring XDG_CONFIG_HOME.
func configPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "genius", "config.toml")
}

// LoadConfig reads the config file if present. A missing file is not an error:
// it returns a zero Config so defaults apply.
func LoadConfig() (Config, error) {
	var c Config
	path := configPath()
	if path == "" {
		return c, nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c, nil
	}
	_, err := toml.DecodeFile(path, &c)
	return c, err
}
