package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// PermConfig is config for perm
type PermConfig struct {
	Debug          bool
	GithubToken    string
	GithubUsername string

	// discouraged from setting, special chars can break
	GithubPassword string
}

// GetConfig gets PermConfig from env
// vars are prefixed with PERM_ and are all caps
func GetConfig() (c PermConfig) {
	err := envconfig.Process("perm", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	return
}
