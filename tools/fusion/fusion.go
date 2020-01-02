package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

var (
	apiKey     = "Kc-ntFUr767Qyk0RgbVU8dxLLdjmTM98_XjsAZxOajA"
	baseURL, _ = url.Parse("http://209.222.106.245:9011/")
	httpClient = &http.Client{Timeout: time.Second * 5}
	auth       = fusionauth.NewClient(httpClient, baseURL, apiKey)
)

// login the fusion
func login() string {
	var credentials fusionauth.LoginRequest
	credentials.LoginId = "alon@traderlinked.com"
	credentials.Password = "XjsAZxOajA"

	// Use FusionAuth Go client to log in the user
	response, _, err := auth.Login(credentials)
	if err != nil {
		panic(err)
	}

	return response.Token
}

func main() {
	// login the fusion to fetch token
	token := login()

	// validate the token
	result, err := auth.ValidateJWT(token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("validate token: %v\n", result)
}
