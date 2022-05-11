package auth

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
	"path/filepath"
	"skema-tool/internal/pkg/console"
	"skema-tool/internal/pkg/io"
	"time"
)

const (
	clientID = "e1b4cf22c78730794f27"
	scope    = "repo, user"
)

type GithubAuth struct {
	deviceCode      string
	userCode        string
	verificationUrl string
	accessToken     string
}

// follow the instructions in https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#device-flow

func (g *GithubAuth) StartAuthProcess() {
	g.requestVerificationCode()
	g.waitForUserAuthComplete()
}

func (g *GithubAuth) AccessToken() string {
	return g.accessToken
}

func (g *GithubAuth) requestVerificationCode() {
	// POST https://github.com/login/device/code
	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte(fmt.Sprintf(`{"client_id":"%s", "scope":"%s"}`, clientID, scope))).
		Post("https://github.com/login/device/code")

	result := map[string]string{}
	json.Unmarshal(resp.Body(), &result)
	g.deviceCode = result["device_code"]
	g.userCode = result["user_code"]
	g.verificationUrl = result["verification_uri"]
}

func (g *GithubAuth) waitForUserAuthComplete() {
	console.Info(fmt.Sprintf("please open url %s and input code: %s", g.verificationUrl, g.userCode))
	console.Info("Wait for github auth ...")

	client := resty.New()
	sleepTime := 2
	maxSleepTime := 32
	for {
		// POST https://github.com/login/oauth/access_token
		resp, _ := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody([]byte(fmt.Sprintf(`{"client_id":"%s", "device_code":"%s", "grant_type":"urn:ietf:params:oauth:grant-type:device_code"}`, clientID, g.deviceCode))).
			Post("https://github.com/login/oauth/access_token")

		if resp.IsSuccess() {
			result := map[string]string{}
			json.Unmarshal(resp.Body(), &result)
			if errMessage, ok := result["error"]; ok {
				if errMessage != "" {
					console.Info(errMessage)
					if errMessage == "slow_down" {
						sleepTime *= 2
						if sleepTime >= maxSleepTime {
							sleepTime = maxSleepTime
						}
						console.Info(fmt.Sprintf("retry in %d seconds", sleepTime))
						time.Sleep(time.Duration(sleepTime) * time.Second)
					}
					continue
				}
			}
			g.accessToken = result["access_token"]
			if g.accessToken == "" {
				continue
			}

			console.Infof("Received github token: %s\n", g.accessToken)
			break
		}
	}
}

// Save Token
func (g *GithubAuth) SaveTokenToFile() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	tokenFilepath := filepath.Join(exePath, "github/token")

	io.SaveToFile(tokenFilepath, []byte(g.accessToken))
	console.Info("token save to " + tokenFilepath)
}

func (g *GithubAuth) GetLocalToken() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	tokenFilepath := filepath.Join(exePath, "github/token")
	data, err := os.ReadFile(tokenFilepath)
	return string(data)
}
