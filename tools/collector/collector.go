package main

import (
	gateway "TL-Proto/gateway"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"google.golang.org/grpc"
)

var (
	address   = flag.String("a", "localhost:15001", "the grpc server address")
	frequency = flag.Int("f", 500, "the collect frequency(ms)")
	count     = flag.Int("c", -1, "the count of records")
)

// login the fusion
func login() string {
	apiKey := "Kc-ntFUr767Qyk0RgbVU8dxLLdjmTM98_XjsAZxOajA"
	baseURL, _ := url.Parse("http://209.222.106.245:9011/")
	httpClient := &http.Client{Timeout: time.Second * 5}
	auth := fusionauth.NewClient(httpClient, baseURL, apiKey)

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

func makeData(i int) []byte {
	type Data struct {
		Name  string
		Email string
	}
	d := Data{
		Name:  fmt.Sprintf("N%04d", i),
		Email: "alonlong@163.com",
	}

	b, _ := json.Marshal(&d)
	return b
}

func main() {
	flag.Parse()

	// login the fusion to fetch the token
	token := login()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := gateway.NewReportServiceClient(conn)

	i := 0
	for {
		i++

		request := gateway.ReportRequest{
			Token: token,
			Data:  makeData(i),
		}
		if _, err := c.Report(context.Background(), &request); err != nil {
			fmt.Printf("report data: %v\n", err)
		}

		if *frequency > 0 {
			time.Sleep(time.Millisecond * time.Duration(*frequency))
		}

		if *count == i {
			break
		}
	}
}
