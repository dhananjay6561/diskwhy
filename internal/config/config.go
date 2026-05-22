package config

import (
	"os"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds every runtime setting for diskwhy. Values are resolved once at
// startup via the precedence chain: CLI flag > env var > ~/.diskwhy.yml > default.
type Config struct {
	// Scan behaviour
	Depth     int
	StaleDays int
	Workers   int

	// Output
	NoColor bool
	JSON    bool
	Verbose bool
	Debug   bool

	// Features
	SkipDocker     bool
	Trash          bool
	GitTimeoutSecs int
}

// DefaultWorkers returns min(runtime.NumCPU(), 8). Disk I/O is the bottleneck;
// beyond 8 workers additional goroutines thrash rather than parallelize.
func DefaultWorkers() int {
	n := runtime.NumCPU()
	if n > 8 {
		return 8
	}
	return n
}

// Load builds a Config by reading from viper's resolved state. Callers in the
// cmd package must bind cobra PersistentFlags to viper before calling Load so
// that the flag > env > file > default precedence is respected.
func Load() (*Config, error) {
	setDefaults()

	viper.SetConfigName(".diskwhy")
	viper.SetConfigType("yaml")
	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(home)
	}

	viper.SetEnvPrefix("DISKWHY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Config file is optional; ignore "not found" silently.
	_ = viper.ReadInConfig()

	// NO_COLOR is the universal terminal standard; it has no DISKWHY_ prefix.
	if os.Getenv("NO_COLOR") != "" {
		viper.Set("no_color", true)
	}

	workers := clamp(viper.GetInt("workers"), 1, 32)

	return &Config{
		Depth:          viper.GetInt("depth"),
		StaleDays:      viper.GetInt("stale_days"),
		Workers:        workers,
		NoColor:        viper.GetBool("no_color"),
		JSON:           viper.GetBool("json"),
		Verbose:        viper.GetBool("verbose"),
		Debug:          viper.GetBool("debug"),
		SkipDocker:     viper.GetBool("skip_docker"),
		Trash:          viper.GetBool("trash"),
		GitTimeoutSecs: viper.GetInt("git_timeout_secs"),
	}, nil
}

// BindFlags binds cobra persistent flags to their viper keys so that CLI flags
// take precedence over env vars and config file values.
func BindFlags(flags *pflag.FlagSet) {
	bindings := []struct {
		flag string
		key  string
	}{
		{"no-color", "no_color"},
		{"json", "json"},
		{"verbose", "verbose"},
		{"debug", "debug"},
	}
	for _, b := range bindings {
		if f := flags.Lookup(b.flag); f != nil {
			_ = viper.BindPFlag(b.key, f)
		}
	}
}

func setDefaults() {
	viper.SetDefault("depth", 3)
	viper.SetDefault("stale_days", 90)
	viper.SetDefault("no_color", false)
	viper.SetDefault("json", false)
	viper.SetDefault("skip_docker", false)
	viper.SetDefault("trash", false)
	viper.SetDefault("workers", DefaultWorkers())
	viper.SetDefault("git_timeout_secs", 30)
	viper.SetDefault("verbose", false)
	viper.SetDefault("debug", false)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
