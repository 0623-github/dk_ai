package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	AIConfigPath = "conf/openai.yaml"
)

// GetConfig 读取并解析YAML配置文件，支持泛型返回指定类型
func GetConfig[T any](ctx context.Context, path string) (T, error) {
	var config T

	absPath, err := filepath.Abs(path)
	if err != nil {
		return config, fmt.Errorf("unable to get absolute path of config file: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return config, fmt.Errorf("config file does not exist: %s", absPath)
	}

	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		if jsonErr := json.Unmarshal(data, &config); jsonErr != nil {
			return config, fmt.Errorf("failed to parse config file: YAML: %w, JSON: %w", err, jsonErr)
		}
	}

	return config, nil
}
