package config

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	katconfig "github.com/katzenpost/authority/voting/server/config"
	"github.com/katzenpost/core/crypto/rand"
)

const (
	DefaultLayers               = 3
	DefaultMinNodesPerLayer     = 2
	defaultTendermintConfigPath = "$HOME/.tendermint/config/config.toml"
	// Note: These values are picked primarily for debugging and need to be changed to something more suitable for a production deployment at some point.
	defaultSendRatePerMinute    = 100
	defaultMu                   = 0.00025
	defaultMuMaxPercentile      = 0.99999
	defaultLambdaP              = 0.00025
	defaultLambdaPMaxPercentile = 0.99999
	defaultLambdaL              = 0.00025
	defaultLambdaLMaxPercentile = 0.99999
	defaultLambdaD              = 0.00025
	defaultLambdaDMaxPercentile = 0.99999
	defaultLambdaM              = 0.00025
	defaultLambdaMMaxPercentile = 0.99999
	absoluteMaxDelay            = 6 * 60 * 60 * 1000 // 6 hours.
)

var DefaultParameters = katconfig.Parameters{
	SendRatePerMinute: defaultSendRatePerMinute,
	Mu:                defaultMu,
	MuMaxDelay:        uint64(math.Min(rand.ExpQuantile(defaultMu, defaultMuMaxPercentile), absoluteMaxDelay)),
	LambdaP:           defaultLambdaP,
	LambdaPMaxDelay:   uint64(rand.ExpQuantile(defaultLambdaP, defaultLambdaPMaxPercentile)),
	LambdaL:           defaultLambdaL,
	LambdaLMaxDelay:   uint64(rand.ExpQuantile(defaultLambdaL, defaultLambdaLMaxPercentile)),
	LambdaD:           defaultLambdaD,
	LambdaDMaxDelay:   uint64(rand.ExpQuantile(defaultLambdaD, defaultLambdaDMaxPercentile)),
	LambdaM:           defaultLambdaM,
	LambdaMMaxDelay:   uint64(rand.ExpQuantile(defaultLambdaM, defaultLambdaMMaxPercentile)),
}

type Config struct {
	TendermintConfigPath string
	DBPath               string
	Layers               int
	MinNodesPerLayer     int
	Parameters           katconfig.Parameters
	Membership           bool
}

func DefaultConfig() (cfg *Config) {
	cfg = new(Config)
	_ = cfg.FixupAndValidate()
	return
}

// FixupAndValidate applies defaults to config entries and validates the
// supplied configuration.  Most people should call one of the Load variants
// instead.
func (c *Config) FixupAndValidate() (err error) {
	if len(c.DBPath) <= 0 {
		c.DBPath = filepath.Join(os.TempDir(), "katzenmint")
	}
	if len(c.TendermintConfigPath) <= 0 {
		c.TendermintConfigPath = defaultTendermintConfigPath
	}
	if c.Layers <= 0 {
		c.Layers = DefaultLayers
	}
	if c.MinNodesPerLayer <= 0 {
		c.MinNodesPerLayer = DefaultMinNodesPerLayer
	}
	if c.Parameters.SendRatePerMinute <= 0 {
		c.Parameters.SendRatePerMinute = DefaultParameters.SendRatePerMinute
	}
	if c.Parameters.Mu <= 0 {
		c.Parameters.Mu = DefaultParameters.Mu
	}
	if c.Parameters.MuMaxDelay <= 0 {
		c.Parameters.MuMaxDelay = DefaultParameters.MuMaxDelay
	}
	if c.Parameters.LambdaP <= 0 {
		c.Parameters.LambdaP = DefaultParameters.LambdaP
	}
	if c.Parameters.LambdaPMaxDelay <= 0 {
		c.Parameters.LambdaPMaxDelay = DefaultParameters.LambdaPMaxDelay
	}
	if c.Parameters.LambdaL <= 0 {
		c.Parameters.LambdaL = DefaultParameters.LambdaL
	}
	if c.Parameters.LambdaLMaxDelay <= 0 {
		c.Parameters.LambdaLMaxDelay = DefaultParameters.LambdaLMaxDelay
	}
	if c.Parameters.LambdaD <= 0 {
		c.Parameters.LambdaD = DefaultParameters.LambdaD
	}
	if c.Parameters.LambdaDMaxDelay <= 0 {
		c.Parameters.LambdaDMaxDelay = DefaultParameters.LambdaDMaxDelay
	}
	if c.Parameters.LambdaM <= 0 {
		c.Parameters.LambdaM = DefaultParameters.LambdaM
	}
	if c.Parameters.LambdaMMaxDelay <= 0 {
		c.Parameters.LambdaMMaxDelay = DefaultParameters.LambdaMMaxDelay
	}
	return
}

// Load parses and validates the provided buffer b as a config file body and
// returns the Config.
func Load(b []byte) (*Config, error) {
	cfg := new(Config)
	md, err := toml.Decode(string(b), cfg)
	if err != nil {
		return nil, err
	}
	if undecoded := md.Undecoded(); len(undecoded) != 0 {
		return nil, fmt.Errorf("config: Undecoded keys in config file: %v", undecoded)
	}
	if err := cfg.FixupAndValidate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFile loads, parses, and validates the provided file and returns the
// Config.
func LoadFile(f string) (*Config, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return Load(b)
}

// SaveFile saves the config to the provided file
func SaveFile(f string, config *Config) error {
	file, err := os.Create(f)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := toml.NewEncoder(file)
	return enc.Encode(config)
}
