package main

import (
	gateway "TL-Gateway/proto/gateway"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"google.golang.org/grpc"
)

var (
	address     = flag.String("a", "192.168.0.4:15001", "the grpc server address")
	frequency   = flag.Int("f", 500, "the collect frequency(ms)")
	count       = flag.Int("c", -1, "the count of records")
	connections = flag.Int("s", 1, "the count of connections")
	dataType    = flag.String("t", "data_trace", "the data type(data_trace/data_quote)")
	action      = flag.String("n", "insert", "the action of message(insert/update)")
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
//         "source": "fund-falconfund",
//         "path": "cpu.pct",
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

type QuoteData struct {
	Ticker string  `json:"ticker"`
	Last   float32 `json:"last"`
	Bid    float32 `json:"bid"`
	Ask    float32 `json:"ask"`
	Time   string  `json:"time"`
}

type NetworkQuote struct {
	UserID string    `json:"userid"`
	Source string    `json:"source"`
	Path   string    `json:"path"`
	Kind   string    `json:"dtype"`
	Action string    `json:"action"`
	Data   QuoteData `json:"data"`
}

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
	Action string    `json:"action"`
	Data   TraceData `json:"data"`
	Time   string    `json:"time"`
}

func newNetworkQuote(i int, kind string, action string) []byte {
	quote := &NetworkQuote{
		Kind:   kind,
		Action: action,
		UserID: "test@live.cn",
		Source: "fund-falconfund-" + strconv.Itoa(i),
		Path:   "trade.network.latency",
		Data: QuoteData{
			Ticker: "GBPNZD",
			Last:   1.9717,
			Bid:    1.9718,
			Ask:    1.9719,
			Time:   time.Now().Format("2006-02-01 15:04:05 +0800 CST"),
		},
	}

	b, _ := json.Marshal(&quote)
	return b
}

func newNetworkTrace(i int, kind string, action string) []byte {
	trace := &NetworkTrace{
		Kind:   kind,
		Action: action,
		Source: "fund-falconfund-" + strconv.Itoa(i),
		Path:   "trade.network.latency",
		Data: TraceData{
			Type:    "ipv4",
			Trace:   "alpari.london.trade.api -w 2000",
			Hop:     true,
			Host:    "alpari.london.trade.api [217.192.86.32]",
			Latency: 10,
			Value:   2.997,
		},
		Time: time.Now().Format("2006-02-01 15:04:05 +0800 CST"),
	}

	b, _ := json.Marshal(&trace)
	return b
}

func do(token string, dtype string, action string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(*address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := gateway.NewServiceClient(conn)

	i := 0
	for {
		i++

		var data []byte
		if dtype == "data_trace" {
			data = newNetworkTrace(i, dtype, action)
		}
		if dtype == "data_quote" {
			data = newNetworkQuote(i, dtype, action)
		}

		request := gateway.ReportRequest{
			Token:         token,
			LoginId:       "alon@traderlinked.com",
			ApplicationId: "8af9b71b-5637-435c-9f4c-fb82e17dd114",
			Data:          data,
		}
		if _, err := c.Report(context.Background(), &request); err != nil {
			fmt.Printf("report data: %v\n", err)
			return
		}

		if *frequency > 0 {
			time.Sleep(time.Millisecond * time.Duration(*frequency))
		}

		if *count == i {
			break
		}
	}
}

func main() {
	flag.Parse()

	if *dataType != "data_quote" && *dataType != "data_trace" {
		panic("invalid data type")
	}

	// login the fusion to fetch the token
	token := login()

	for index := 0; index < *connections; index++ {
		go do(token, *dataType, *action)
	}

	sig := make(chan os.Signal, 1024)
	// subscribe signals: SIGINT & SINGTERM
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sig:
			return
		}
	}
}
