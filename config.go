package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	ProxyLinks []ProxyLink `json:"hosts"`
	Settings   Settings    `json:"settings"`
}

type ProxyLink struct {
	LocalAddr  string `json:"local_address"`
	RemoteAddr string `json:"remote_address"`
}

type Settings struct {
	Verbose                bool   `json:"verbose"`      // display server actions
	VeryVerbose            bool   `json:"very_verbose"` // display server actions and all tcp data
	OutputHex              bool   `json:"output_hex"`
	OutputAnsiColors       bool   `json:"output_ansi_colors"`
	DisableNaglesAlgorithm bool   `json:"disable_nagles_algorithm"`
	UnwrapTls              bool   `json:"unwrap_tls"`
	MatchRegex             string `json:"match_regex"`
	ReplaceRegex           string `json:"replace_regex"`
}

func ReadConfig() *Config {
	fileContents, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(fmt.Errorf("failed to read config json file: %w", err))
	}

	var config Config
	if err := json.Unmarshal(fileContents, &config); err != nil {
		panic(fmt.Errorf("failed to parse config json file: %w", err))
	}

	if len(config.ProxyLinks) < 1 {
		panic(fmt.Errorf("must specify at least one local to remote host link"))
	} else if len(config.ProxyLinks) > 1 {
		panic(fmt.Errorf("multiple local to remote host links are not supported"))
	}

	return &config
}
