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
	"strings"
	"sync"
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
	loginid     = flag.String("u", "alonlong@126.com", "the username for fusion")
	password    = flag.String("p", "1qaz@wsx", "the password for fusion")
)

// login the fusion
func login(loginid string, password string) (string, string) {
	apiKey := "gpxAG9Yk1_3xkItITC35rXP2zVdbsdk_bep69TWkGC8"
	baseURL, _ := url.Parse("http://209.159.148.254:9011/")
	httpClient := &http.Client{Timeout: time.Second * 5}
	auth := fusionauth.NewClient(httpClient, baseURL, apiKey)

	var credentials fusionauth.LoginRequest
	credentials.LoginId = loginid
	credentials.Password = password

	// Use FusionAuth Go client to log in the user
	response, _, err := auth.Login(credentials)
	if err != nil {
		panic(err)
	}

	return response.Token, response.User.Id
}

type QuoteData struct {
	Ticker string  `json:"ticker"`
	Last   float32 `json:"last"`
	Bid    float32 `json:"bid"`
	Ask    float32 `json:"ask"`
	Time   string  `json:"time"`
}

type NetworkQuote struct {
	UserID int       `json:"userid"`
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

func newNetworkQuote(i int, userid int, kind string, action string) []byte {
	quote := &NetworkQuote{
		Kind:   kind,
		Action: action,
		UserID: userid,
		Source: "fund-falconfund-" + strconv.Itoa(i),
		Path:   "trade.network.latency",
		Data: QuoteData{
			Ticker: "GBPNZD",
			Last:   1.9717,
			Bid:    1.9718,
			Ask:    1.9719,
			Time:   time.Now().Format("2006-01-02 15:04:05"),
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
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}

	b, _ := json.Marshal(&trace)
	return b
}

func do(wg *sync.WaitGroup, token string, userid int, dtype string, action string) {
	defer wg.Done()

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
			data = newNetworkTrace(0, dtype, action)
		}
		if dtype == "data_quote" {
			data = newNetworkQuote(0, userid, dtype, action)
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

// dbd62208-d0ea-4a5a-9066-64d41e7dcedd
// 8+4+4+4+12
// for the uuid format, i will put the user id with 10 bytes replace the part "d0ea-4a5a-90"
func uuid2id(uuid string) int {
	if len(uuid) != 36 {
		panic(uuid)
	}

	out := strings.Replace(uuid[9:21], "-", "", -1)
	id, _ := strconv.Atoi(out)
	return id
}

func main() {
	flag.Parse()

	if *dataType != "data_quote" && *dataType != "data_trace" {
		panic("invalid data type")
	}

	// login the fusion to fetch the token
	token, uuid := login(*loginid, *password)

	userid := uuid2id(uuid)

	var wg sync.WaitGroup
	for index := 0; index < *connections; index++ {
		wg.Add(1)
		go do(&wg, token, userid, *dataType, *action)
	}

	wg.Wait()
}
