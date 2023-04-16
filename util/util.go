package util

import (
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
	valStr := GetEnv(varName, defaultVal)
	defaultBool, _ := strconv.ParseBool(defaultVal)

	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false
	}

	// edge case if the pointer is unset
	if ptr == nil {
		return val
	}

	// when the pointer value is not the same as the default value, use it
	if *ptr != defaultBool {
		return *ptr
	} else {
		// when the pointer and the default value are the same, use the result of getenv
		return val
	}
}

// GetEnv looks up an environment with a fallback default
func GetEnv(varName, defaultVal string) string {
	if value, isPresent := os.LookupEnv(varName); isPresent {
		return value
	}

	return defaultVal
}

// IPMonitored considers configured allowed and blocked ip addresses and inspects a clients
// public ip address to determine if it should be used for speed and connectivity testing
func IPMonitored(publicIP string, allowed, blocked []string) bool {
	return ipAllowed(publicIP, allowed) && !ipBlocked(publicIP, blocked)
}

// iterate over list of allowed ips and ensure the public ip is a match
func ipAllowed(publicIP string, ips []string) bool {
	allowed := true
	for _, v := range ips {
		if v == "" {
			continue
		}

		if publicIP == v {
			allowed = true
			return true
		}

		allowed = false
	}

	return allowed
}

// iterate over list of blocked ips and ensure the public ip is not a match
func ipBlocked(publicIP string, ips []string) bool {
	blocked := false

	for _, v := range ips {
		if v == "" {
			continue
		}

		if publicIP == v {
			blocked = true
			break
		}

		blocked = false
	}

	return blocked
}
