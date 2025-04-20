package hz_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var InvalidKey = errors.New("invalid configuration key specified")

func getEnv(key string) (string, error) {
	val, set := os.LookupEnv(key)
	if !set {
		return "", InvalidKey
	}
	return val, nil
}

type Config struct {
	path   string
	bytes  []byte
	values interface{}
}

func NewConfigFromPath(path string) *Config {
	return &Config{
		path: path,
	}
}

func NewConfigFromString(value string) *Config {
	return &Config{
		bytes: []byte(value),
	}
}

func NewConfigWithDefaultPath(environment string) *Config {
	return NewConfigFromPath(fmt.Sprintf("./config/%s.json", strings.ToLower(environment)))
}

func NewConfigWithEnvironmentVariableName(name string) *Config {
	env, err := getEnv(name)
	if err != nil {
		env = "dev"
	}
	return NewConfigWithDefaultPath(env)
}

func NewConfigWithAutomaticDetection() *Config {
	return NewConfigWithEnvironmentVariableName("ENVIRONMENT")
}

func (c *Config) Read() error {
	bytes, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	c.bytes = bytes
	return c.Parse()
}

func (c *Config) Parse() error {
	var data interface{}
	err := json.Unmarshal(c.bytes, &data)
	if err != nil {
		return err
	}
	c.values = data
	c.bytes = nil // throw away bytes after reading them
	return nil
}

func (c *Config) ValueOrPanic(key string) string {
	v, err := c.Value(key)
	if err != nil {
		panic("Could not find value for key: " + key)
	}
	return v
}

func (c *Config) IntValueOrPanic(key string) int {
	v := c.ValueOrPanic(key)
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic("could not convert key: " + key + " with value: " + v + " to int")
	}
	return int(f)
}

func (c *Config) BoolValueOrPanic(key string) bool {
	v := c.ValueOrPanic(key)
	f, err := strconv.ParseBool(v)
	if err != nil {
		panic("could not convert key: " + key + " with value: " + v + " to bool")
	}
	return f
}

func (c *Config) StringSliceValueOrPanic(key string) []string {
	ss, err := c.StringSliceValue(key)
	if err != nil {
		panic("could not convert key: " + key + " to []string")
	}
	return ss
}

func (c *Config) StringSliceValue(key string) ([]string, error) {
	value, ok := c.getValue(key)
	if !ok {
		return nil, InvalidKey
	}

	var ss []string
	if anySlice, ok := value.([]any); ok {
		for _, val := range anySlice {
			valStr, ok := convertToString(val)
			if !ok {
				return nil, InvalidKey
			}
			valStr = strings.TrimSpace(valStr)

			resolvedValue, ok := resolveValue(valStr)
			if !ok {
				return nil, InvalidKey
			}
			ss = append(ss, resolvedValue)
		}
	}

	if ss != nil {
		return ss, nil
	}

	return nil, InvalidKey
}

func (c *Config) Value(key string) (string, error) {
	value, ok := c.getValue(key)
	if !ok {
		return "", InvalidKey
	}

	valueStr, ok := convertToString(value)
	if !ok {
		return "", InvalidKey
	}

	resolved, ok := resolveValue(valueStr)
	if !ok {
		return "", InvalidKey
	}

	return resolved, nil
}

func (c *Config) getValue(key string) (any, bool) {
	obj, ok := c.values.(map[string]interface{})
	if !ok {
		return "", false
	}

	parts := strings.Split(key, ".")
	for _, part := range parts[:len(parts)-1] {
		i, ok := obj[part]
		if !ok {
			return "", false
		}
		obj, ok = i.(map[string]interface{})
		if !ok {
			return "", false
		}
	}

	value, ok := obj[parts[len(parts)-1]]
	if !ok {
		return "", false
	}

	return value, true
}

func convertToString(value any) (string, bool) {
	if valueStr, ok := value.(string); ok {
		return valueStr, true
	}

	if intStr, ok := value.(int); ok {
		return strconv.Itoa(intStr), true
	}

	if floatStr, ok := value.(float64); ok {
		return fmt.Sprintf("%f", floatStr), true
	}

	if boolStr, ok := value.(bool); ok {
		return strconv.FormatBool(boolStr), true
	}

	return "", false
}

func resolveValue(valueStr string) (string, bool) {
	n := len(valueStr)
	if n > 4 {
		if valueStr[0:2] == "{{" && valueStr[n-2:n] == "}}" {
			var err error
			valueStr, err = getEnv(valueStr[2 : n-2])
			if err != nil {
				return "", false
			}
		}
	}
	return valueStr, true
}
