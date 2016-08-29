package main

import (
	"log"

	"github.com/mozilla-services/perm/config"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	conf   config.PermConfig
	client *github.Client
)

func main() {
	conf = config.GetConfig()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: conf.GithubToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client = github.NewClient(tc)

	if conf.Debug {
		user, _, err := client.Users.Get("")
		if err != nil {
			log.Fatalf("[error] github: %s", err.Error())
		}
		log.Printf("[debug] authenticated as user: %s", *user.Login)
	}
}
