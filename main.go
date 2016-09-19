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
)

func init() {
	log.SetPrefix("perm: ")
}

func main() {
	var (
		otp string
	)

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

	log.Printf("[info] Auditing authorizations for user %q", conf.GithubUsername)
	auditAuthorizations(otp)

	log.Printf("[info] Auditing grants for user %q", conf.GithubUsername)
	auditGrants(otp)
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

func makeReq(url, otp string) (*http.Request, error) {
	req, err := client.NewRequest(http.MethodGet, "/authorizations", nil)
	if err != nil {
		return nil, err
	}

	if len(otp) > 0 {
		// OTP header
		req.Header.Set("X-Github-OTP", otp)
	}

	// required for preview API, see: https://developer.github.com/v3/oauth_authorizations/
	req.Header.Set("Accept", "application/vnd.github.damage-preview")
	req.SetBasicAuth(conf.GithubUsername, conf.GithubPassword)
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

func auditAuthorizations(otp string) {
	req, err := makeReq("/authorizations", otp)
	if err != nil {
		log.Fatalf("[error] could not make authorizations request: %v", err.Error())
	}
	var authorizations []*github.Authorization
	resp, err := client.Do(req, &authorizations)
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("[error] Do: %v, status %s", err.Error(), resp.Status)
	}
	for _, auth := range authorizations {
		hasRelevantScope := false
		for _, scope := range auth.Scopes {
			if relevant := isScopeRelevant(string(scope)); relevant {
				hasRelevantScope = true
			}
		}
		if !hasRelevantScope {
			if conf.Debug {
				log.Printf("[debug] authorization for %q has no relevant scope (has %v)", *auth.App.Name, auth.Scopes)
			}
			continue
		}
		log.Printf("[info] authorization for %q, (%d) was created: %s, last updated: %s, has scopes: %v\n",
			*auth.App.Name,
			*auth.ID,
			auth.CreatedAt.Format("02 Jan 06"),
			auth.UpdateAt.Format("02 Jan 06"),
			auth.Scopes,
		)
	}
}

func auditGrants(otp string) {
	req, err := makeReq("/applications/grants", otp)
	if err != nil {
		log.Fatalf("[error] could not make grants request: %v", err.Error())
	}
	var grants []*github.Grant
	resp, err := client.Do(req, &grants)
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("[error] Do: %v, status %s", err.Error(), resp.Status)
	}
	for _, grant := range grants {
		hasRelevantScope := false
		for _, scope := range grant.Scopes {
			if relevant := isScopeRelevant(scope); relevant {
				hasRelevantScope = true
			}
		}
		if !hasRelevantScope {
			if conf.Debug {
				log.Printf("[debug] grant for %q has no relevant scope (has %v)", *grant.App.Name, grant.Scopes)
			}
			continue
		}
		log.Printf("[info] grant for %q (%d) was created: %s, last updated: %s, has scopes: %v\n",
			*grant.App.Name,
			*grant.ID,
			grant.CreatedAt.Format("02 Jan 06"),
			grant.UpdatedAt.Format("02 Jan 06"),
			grant.Scopes,
		)
	}
}
