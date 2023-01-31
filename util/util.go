package util

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
)

// ValueOr returns a de-referenced string pointer, an environment variable, or a fallback
func ValueOr(ptr *string, varName, defaultVal string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}

	return GetEnv(varName, defaultVal)
}

// BooleanValueOr returns a de-referenced boolean pointer, an environment variable, or a fallback
func BooleanValueOr(ptr *bool, varName, defaultVal string) bool {
	if ptr != nil {
		return *ptr
	}

	valStr := GetEnv(varName, defaultVal)
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false
	}
	return val
}

// GetEnv looks up an environment with a fallback default
func GetEnv(varName, defaultVal string) string {
	if value, isPresent := os.LookupEnv(varName); isPresent {
		return value
	}

	return defaultVal
}

func PublicIP() (string, error) {
	req, err := http.Get("https: //api.ipify.org?format=json/")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	type IP struct {
		IP string `json:"ip"`
	}
	var ip IP
	if err := json.Unmarshal(body, &ip); err != nil {
		return "", err
	}

	return ip.IP, nil
}
