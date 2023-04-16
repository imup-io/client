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
	val, err := strconv.ParseBool(GetEnv(varName, defaultVal))
	if err != nil {
		return false
	}

	if ptr == nil {
		return val
	}

	return val || *ptr
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
