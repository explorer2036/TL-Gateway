package test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

const host = "http://209.222.106.245:9011/"

var apiKey = "Kc-ntFUr767Qyk0RgbVU8dxLLdjmTM98_XjsAZxOajA"
var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

var baseURL, _ = url.Parse(host)

// Construct a new FusionAuth Client
var auth = fusionauth.NewClient(httpClient, baseURL, apiKey)

func TestFusionLogin(t *testing.T) {
	var credentials fusionauth.LoginRequest
	credentials.ApplicationId = ""
	credentials.LoginId = "alon@traderlinked.com"
	credentials.Password = "XjsAZxOajA"

	// Use FusionAuth Go client to log in the user
	response, _, err := auth.Login(credentials)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("login token: %v\n", response.Token)

	result, err := auth.ValidateJWT(response.Token)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("validate token: %v\n", result)
}
