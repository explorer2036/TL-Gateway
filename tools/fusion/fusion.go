package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

var (
	// apiKey = "Kc-ntFUr767Qyk0RgbVU8dxLLdjmTM98_XjsAZxOajA"
	apiKey     = "cIoPv09UNR2qhd4OrHDbq1x1qeHpZpumUMvIAuWO0no"
	baseURL, _ = url.Parse("http://209.222.106.245:9011/")
	httpClient = &http.Client{Timeout: time.Second * 5}
	auth       = fusionauth.NewClient(httpClient, baseURL, apiKey)
)

func struct2JSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// login the fusion
func login() (string, string) {
	var credentials fusionauth.LoginRequest
	credentials.LoginId = "alon@traderlinked.com"
	credentials.Password = "XjsAZxOajA"
	credentials.ApplicationId = "8af9b71b-5637-435c-9f4c-fb82e17dd114"
	credentials.IpAddress = ""

	// Use FusionAuth Go client to log in the user
	response, _, err := auth.Login(credentials)
	if err != nil {
		panic(err)
	}

	return response.Token, response.RefreshToken
}

// refresh the token
func refresh(token string) string {
	var refreshRequest fusionauth.RefreshRequest
	refreshRequest.RefreshToken = token

	response, _, err := auth.ExchangeRefreshTokenForJWT(refreshRequest)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(response.RefreshTokens))
	fmt.Printf("response: %v\n", struct2JSON(response))

	return response.Token
}

func main() {
	// login the fusion to fetch token
	token, refreshToken := login()

	fmt.Printf("token: %v\n", token)
	fmt.Printf("refresh token: %v\n", refreshToken)

	// validate the token
	result, err := auth.ValidateJWT(token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("validate token: %v\n", result)

	// refresh the token
	token = refresh(refreshToken)
	fmt.Printf("token: %v\n", token)
	// validate the token
	result, err = auth.ValidateJWT(token)
	if err != nil {
		panic(err)
	}
	fmt.Printf("validate token: %v\n", result)

}
