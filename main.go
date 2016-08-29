package main

import (
	"fmt"
	"log"
	"net/http"

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

	tok := oauth2.Token{AccessToken: conf.GithubToken}
	ts := oauth2.StaticTokenSource(
		&tok,
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client = github.NewClient(tc)

	user, _, err := client.Users.Get("")
	if err != nil {
		log.Fatalf("[error] github: %s", err.Error())
	}
	if conf.Debug {
		log.Printf("[debug] authenticated as user: %s", *user.Login)
	}

	auditUser(user)
}

func auditUser(user *github.User) {
	// get user's authorizations
	// opt := &github.ListOptions{PerPage: 100}
	// var allAuths []*github.Authorization
	// for {
	// 	auths, resp, err := client.Authorizations.List(opt)
	// 	if err != nil {
	// 		log.Fatalf("[error] listing authorizations for %s: %s", *user.Login, err.Error())
	// 	}
	// 	if resp.NextPage == 0 {
	// 		break
	// 	}
	// 	allAuths = append(allAuths, auths...)
	// 	opt.Page = resp.NextPage
	// }

	// // get user's grants
	req, err := client.NewRequest(http.MethodGet, "/applications/grants", nil)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}
	req.SetBasicAuth(*user.Login, conf.Password)

	var otp string
	log.Println("Enter GitHub 2FA code if applicable")
	n, err := fmt.Scanln(&otp)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}
	if conf.Debug {
		log.Printf("[debug] Got OTP %s", otp)
	}
	if n > 0 {
		req.Header.Set("X-Github-OTP", otp)
	}
	// req.Header.Set("Accept", "application/vnd.github.damage-preview")
	var grants []*github.Grant
	_, err = client.Do(req, &grants)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}
	log.Println(grants)
	// grants, _, err := client.Authorizations.ListGrants()
	// if err != nil {
	// 	log.Fatalf("[error] listing grants for %s: %s", *user.Login, err.Error())
	// }
	// for _, grant := range grants {
	// 	log.Printf("[info] mcrabill has a grant for %s", *grant.App.Name)
	// }
}
