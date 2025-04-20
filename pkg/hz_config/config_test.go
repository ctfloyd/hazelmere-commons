package hz_config

import (
	"os"
	"testing"
)

func TestConfig_StringSliceOrPanic(t *testing.T) {
	err := os.Setenv("ENV_TOKEN", "test_value")
	if err != nil {
		t.Error(err)
	}

	config := NewConfigFromString(`{ "tokens": ["{{ENV_TOKEN}}",     "SECRET_TOKEN"] }`)

	err = config.Parse()
	if err != nil {
		t.Error(err)
	}

	ss, err := config.StringSliceValue("tokens")
	if err != nil {
		t.Error(err)
	}

	if ss[0] != "test_value" {
		t.Error("Expected test_value, got", ss[0])
	}

	if ss[1] != "SECRET_TOKEN" {
		t.Error("Expected SECRET_TOKEN, got", ss[1])
	}
}
