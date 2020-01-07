package main

import (
	gateway "TL-Gateway/proto/gateway"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"google.golang.org/grpc"
)

var (
	address   = flag.String("a", "192.168.0.4:15001", "the grpc server address")
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

// Metric:
//     {
//         "source": "fund-falconfund-nyc-mt4-5",
//         "path": "alpari-brokerage.metatrader.app.cpu.pct",
//         "data": {value: [0.09]},
//         "time": 15990932023
//     }

// CREATE TABLE trace (
// 	source VARCHAR(100) NOT NULL,
// 	path VARCHAR(200) NOT NULL,
// 	type VARCHAR(20) NOT NULL,
// 	trace VARCHAR(100) NOT NULL,
// 	hop BOOLEAN NOT NULL,
// 	host VARCHAR(100) NOT NULL,
// 	latency INT NOT NULL,
// 	value FLOAT(2) NOT NULL,
// 	time TIMESTAMPTZ  NOT NULL,
// );

type TraceData struct {
	Type    string  `json:"type"`
	Trace   string  `json:"trace"`
	Hop     bool    `json:"hop"`
	Host    string  `json:"host"`
	Latency int     `json:"latency"`
	Value   float32 `json:"value"`
}
type NetworkTrace struct {
	Source string    `json:"source"`
	Path   string    `json:"path"`
	Kind   string    `json:"dtype"`
	Data   TraceData `json:"data"`
	Time   string    `json:"time"`
}

func newNetworkTrace(i int) []byte {
	trace := &NetworkTrace{
		Kind:   "data_trace",
		Source: "fund-falconfund-nyc-mt4-5-" + strconv.Itoa(i),
		Path:   "alpari-brokerage.metatrader.app.trade.network.latency",
		Data: TraceData{
			Type:    "ipv4",
			Trace:   "alpari.london.trade.api -w 2000",
			Hop:     true,
			Host:    "alpari.london.trade.api [217.192.86.32]",
			Latency: 10,
			Value:   2.99792458e8,
		},
		// Time: "2015-01-12 12:45:00.345+08",
		Time: "2015-01-12 12:45:01",
	}

	b, _ := json.Marshal(&trace)
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
			Data:  newNetworkTrace(i),
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
