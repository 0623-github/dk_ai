package helper

import (
	"context"
	"testing"
)

func TestConfigExample(t *testing.T) {
	type Config struct {
		Model   string `yaml:"model"`
		BaseURL string `yaml:"baseURL"`
	}
	config, err := GetConfig[Config](context.Background(), "../../conf/openai.yaml")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if config.Model == "" {
		t.Fatalf("Model is empty")
	}
	if config.BaseURL == "" {
		t.Fatalf("BaseURL is empty")
	}
	t.Logf("Model: %s, BaseURL: %s", config.Model, config.BaseURL)
}
