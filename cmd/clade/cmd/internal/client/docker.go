package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

var DefaultDockerConfigPath = ""

type dockerConfig struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}

func LoadAuthFromDockerConfig(path string) (map[string]BasicAuth, error) {
	if path == "" {
		return make(map[string]BasicAuth), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open: %w", err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}

	var conf dockerConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse Docker config: %w", err)
	}

	rst := make(map[string]BasicAuth)
	for svc, entry := range conf.Auths {
		auth_raw, err := base64.StdEncoding.DecodeString(entry.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to decode %s: %w", svc, err)
		}

		auth := strings.SplitN(string(auth_raw), ":", 2)
		if len(auth) != 2 || auth[0] == "" || auth[1] == "" {
			return nil, fmt.Errorf("invalid auth for %s", svc)
		}

		rst[svc] = BasicAuth{
			Username: auth[0],
			Password: auth[1],
		}
	}

	return rst, nil
}

func init() {
	if home, err := os.UserHomeDir(); err == nil {
		DefaultDockerConfigPath = path.Join(home, ".docker", "config.json")
	}
}
