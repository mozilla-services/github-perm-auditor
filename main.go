package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/howeyc/gopass"
	"go.mozilla.org/github-perm-auditor/config"
)

var (
	conf   config.PermConfig
	client *github.Client
	otp    string
)

func init() {
	log.SetPrefix("perm: ")
}

func main() {
	// envconfig
	conf = config.GetConfig()

	// get username
	if conf.GithubUsername == "" {
		log.Println("Enter GitHub username")
		conf.GithubUsername = scan()
	}

	// get password
	if conf.GithubPassword == "" {
		log.Println("Enter GitHub password")
		pass, err := gopass.GetPasswd()
		if err != nil {
			log.Fatalf("[error] could not read password: %v", err.Error())
		}
		conf.GithubPassword = string(pass)
	}

	// get OTP code from stdin
	log.Println("Enter GitHub 2FA code if applicable")
	otp = scan()

	client = github.NewClient(&http.Client{})

	auditAuthorizations()
	auditGrants()
}

func scan() string {
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Fatalf("[error] Scanln: %s", err.Error())
	}
	if conf.Debug {
		log.Printf("[debug] Read value %s", input)
	}
	return input
}

// MakeReq creates an *http.Requestwith the required credentials for the Github Authorizations API
func MakeReq(url, username, password, otp string) (*http.Request, error) {
	req, err := client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if len(otp) > 0 {
		// OTP header
		req.Header.Set("X-Github-OTP", otp)
	}

	// required for preview API, see: https://developer.github.com/v3/oauth_authorizations/
	req.Header.Set("Accept", "application/vnd.github.damage-preview")
	req.SetBasicAuth(username, password)
	return req, err
}

func isScopeRelevant(scope string) bool {
	switch github.Scope(scope) {
	case github.ScopeUserEmail,
		github.ScopeUserFollow,
		github.ScopeNone,
		github.ScopeNotifications,
		github.ScopeGist,
		github.ScopeReadPublicKey,
		github.ScopeReadGPGKey:
		return false
	}
	return true
}

func auditAuthorizations() {
	req, err := MakeReq("/authorizations", conf.GithubUsername, conf.GithubPassword, otp)
	if err != nil {
		log.Fatalf("[error] could not make authorizations request: %v", err.Error())
	}
	var authorizations []*github.Authorization
	resp, err := client.Do(req, &authorizations)
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("[error] Do: %v, status %s", err.Error(), resp.Status)
	}

	log.Printf("[info] Auditing authorizations for user %q", conf.GithubUsername)
	fmt.Printf("Authorizations:\n")

	authsByScope := make(map[github.Scope][]*github.Authorization)
	for _, auth := range authorizations {
		for _, scope := range auth.Scopes {
			authsByScope[scope] = append(authsByScope[scope], auth)
		}
	}

	for scope, auths := range authsByScope {
		if relevant := isScopeRelevant(string(scope)); !relevant {
			if conf.Debug {
				log.Printf("[debug] skipping scope %s", scope)
			}
			continue
		}
		fmt.Printf("  scope: %s\n", scope)
		for _, auth := range auths {
			fmt.Printf("      - authorization for %q: created %s, last updated %s\n",
				*auth.App.Name,
				auth.CreatedAt.Format("02 Jan 06"),
				auth.UpdateAt.Format("02 Jan 06"),
			)
		}
	}
}

func auditGrants() {
	req, err := MakeReq("/applications/grants", conf.GithubUsername, conf.GithubPassword, otp)
	if err != nil {
		log.Fatalf("[error] could not make grants request: %v", err.Error())
	}
	var grants []*github.Grant
	resp, err := client.Do(req, &grants)
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("[error] Do: %v, status %s", err.Error(), resp.Status)
	}

	log.Printf("[info] Auditing grants for user %q", conf.GithubUsername)
	fmt.Printf("Grants:\n")

	grantsByScope := make(map[github.Scope][]*github.Grant)
	for _, grant := range grants {
		for _, scope := range grant.Scopes {
			grantsByScope[github.Scope(scope)] = append(grantsByScope[github.Scope(scope)], grant)
		}
	}

	for scope, grants := range grantsByScope {
		if relevant := isScopeRelevant(string(scope)); !relevant {
			if conf.Debug {
				log.Printf("[debug] skipping scope %s", scope)
			}
			continue
		}
		fmt.Printf("  scope: %s\n", scope)
		for _, grant := range grants {
			fmt.Printf("      - grant for %q: created %s, last updated %s\n",
				*grant.App.Name,
				grant.CreatedAt.Format("02 Jan 06"),
				grant.UpdatedAt.Format("02 Jan 06"),
			)
		}
	}
}
