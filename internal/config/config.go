package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Hotkeys HotkeyConfig `toml:"hotkeys"`
}

type HotkeyConfig struct {
	Modifier    string `toml:"modifier"`
	ScrollLeft  string `toml:"scroll_left"`
	ScrollRight string `toml:"scroll_right"`
	FocusUp     string `toml:"focus_up"`
	FocusDown   string `toml:"focus_down"`
}

func Default() Config {
	return Config{
		Hotkeys: HotkeyConfig{
			Modifier:    "ctrl+cmd+alt",
			ScrollLeft:  "h",
			ScrollRight: "l",
			FocusUp:     "k",
			FocusDown:   "j",
		},
	}
}

func Load(path string) (Config, error) {
	_, err := os.Open(path)
	if err != nil {
		return Default(), nil
	}

	var config Config
	_, err = toml.DecodeFile(path, &config)
	if err != nil {
		return Default(), err
	}
	return config, nil
}
