package auth

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/skema-dev/skema-tool/internal/pkg/console"
	"github.com/skema-dev/skema-tool/internal/pkg/io"
	"os"
	"path/filepath"
	"time"
)

var (
	ClientID = "e1b4cf22c78730794f27"
	scope    = "repo, user"
)

type githubAuth struct {
	deviceCode      string
	userCode        string
	verificationUrl string
	accessToken     string
}

// follow the instructions in https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#device-flow

func (g *githubAuth) StartAuthProcess() {
	g.requestVerificationCode()
	g.waitForUserAuthComplete()
}

func (g *githubAuth) AccessToken() string {
	return g.accessToken
}

func (g *githubAuth) requestVerificationCode() {
	// POST https://github.com/login/device/code
	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte(fmt.Sprintf(`{"client_id":"%s", "scope":"%s"}`, ClientID, scope))).
		Post("https://github.com/login/device/code")

	result := map[string]string{}
	json.Unmarshal(resp.Body(), &result)
	g.deviceCode = result["device_code"]
	g.userCode = result["user_code"]
	g.verificationUrl = result["verification_uri"]
}

func (g *githubAuth) waitForUserAuthComplete() {
	console.Info(fmt.Sprintf("please open url %s and input code: %s", g.verificationUrl, g.userCode))
	console.Info("After github auth is done, press ENTER key to continue...")
	fmt.Scanln()

	client := resty.New()
	sleepTime := 2
	maxSleepTime := 32
	for {
		// POST https://github.com/login/oauth/access_token
		resp, _ := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody([]byte(fmt.Sprintf(`{"client_id":"%s", "device_code":"%s", "grant_type":"urn:ietf:params:oauth:grant-type:device_code"}`, ClientID, g.deviceCode))).
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
func (g *githubAuth) SaveTokenToFile() {
	homePath := io.GetHomePath()
	tokenFilepath := filepath.Join(homePath, "github/token")

	io.SaveToFile(tokenFilepath, []byte(g.accessToken))
	console.Info("token save to " + tokenFilepath)
}

func (g *githubAuth) GetLocalToken() string {
	homePath := io.GetHomePath()
	tokenFilepath := filepath.Join(homePath, "github/token")
	data, _ := os.ReadFile(tokenFilepath)
	return string(data)
}
