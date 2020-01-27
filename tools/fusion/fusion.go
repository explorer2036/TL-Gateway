package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

var (
	// apiKey = "Kc-ntFUr767Qyk0RgbVU8dxLLdjmTM98_XjsAZxOajA"
	// apiKey     = "cIoPv09UNR2qhd4OrHDbq1x1qeHpZpumUMvIAuWO0no"
	apiKey      = "gpxAG9Yk1_3xkItITC35rXP2zVdbsdk_bep69TWkGC8"
	baseURL, _  = url.Parse("http://209.222.106.245:9011/")
	httpClient  = &http.Client{Timeout: time.Second * 5}
	auth        = fusionauth.NewClient(httpClient, baseURL, apiKey)
	application = "8af9b71b-5637-435c-9f4c-fb82e17dd114"
)

func struct2JSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// login the fusion
func login() (string, string) {
	var credentials fusionauth.LoginRequest
	// credentials.LoginId = "alon@traderlinked.com"
	// credentials.Password = "XjsAZxOajA"
	credentials.LoginId = "alonlong@126.com"
	credentials.Password = "1qaz@wsx"
	credentials.ApplicationId = application
	credentials.IpAddress = ""

	// Use FusionAuth Go client to log in the user
	response, _, err := auth.Login(credentials)
	if err != nil {
		panic(err)
	}
	fmt.Printf("login: %v\n", struct2JSON(response))

	return response.Token, response.RefreshToken
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// dbd62208-d0ea-4a5a-9066-64d41e7dcedd
// 8+4+4+4+12
// for the uuid format, i 'll put the user id with 10 bytes replace the part "d0ea-4a5a-90"
func id2uuid(id int) string {
	s := fmt.Sprintf("%010v", id)
	// return fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
	uuid := fmt.Sprintf("%08v-%s-%s-%s%02v-%012v",
		rand.Int31n(100000000),
		s[0:4],
		s[4:8],
		s[8:10],
		rand.Int31n(100),
		rand.Int63n(1000000000000),
	)
	return uuid
}

func register(id int) {

}

// create the user from fusion
func create(id int) {
	// 8af9b71b-5637-435c-9f4c-fb82e17dd114
	uuid := id2uuid(id)
	fmt.Println(uuid)

	registrationRequest := fusionauth.RegistrationRequest{
		GenerateAuthenticationToken:  true,
		SendSetPasswordEmail:         false,
		SkipRegistrationVerification: true,
		Registration: fusionauth.UserRegistration{
			ApplicationId: application,
		},
		User: fusionauth.User{
			SecureIdentity: fusionauth.SecureIdentity{
				Password: "1qaz@wsx",
			},
			Email: "alonlong@126.com",
		},
	}
	registrationReply, _, err := auth.Register(uuid, registrationRequest)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", struct2JSON(registrationReply))
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
	// create(100000010)

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

	// // refresh the token
	// token = refresh(refreshToken)
	// fmt.Printf("token: %v\n", token)
	// // validate the token
	// result, err = auth.ValidateJWT(token)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("validate token: %v\n", result)

}
