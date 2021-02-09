package config

import "strings"

// Env is abstracted environment
type Env struct {
	m map[string]string
}

// Get an environment variable by key, or blank string if missing
func (e *Env) Get(key string) string {
	key = strings.ReplaceAll(key, "localEnv:", "")
	value, ok := e.m[key]
	if !ok {
		return ""
	}
	return value
}

// Set adds an environment variable
func (e *Env) Set(key, value string) {
	e.m[key] = value
}

// NewEnv creates a new env from = separated string slice (eg: os.Environ())
func NewEnv(environ []string) Env {
	e := make(map[string]string)
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		e[parts[0]] = parts[1]
	}
	return Env{m: e}
}
