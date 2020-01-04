package main

import (
	"TL-Gateway/config"
	"TL-Gateway/kafka"
	"TL-Gateway/proto/gateway"
	"TL-Gateway/server"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"google.golang.org/grpc"
)

const (
	// defaultFustionTimeout defines the default timeout for http request every time
	defaultFustionTimeout = 5

	// Success for grpc requests
	Success = 0
)

// ReportService is used to implement gateway.ReportService.
type ReportService struct {
	gateway.UnimplementedReportServiceServer

	producer     *kafka.Producer
	fusionClient *fusionauth.FusionAuthClient
	settings     *config.Config
}

// new a fusion client
func newFusionClient(settings *config.Config) *fusionauth.FusionAuthClient {
	if settings.Fusion.Timeout == 0 {
		settings.Fusion.Timeout = defaultFustionTimeout
	}

	client := &http.Client{
		Timeout: settings.Fusion.Timeout * time.Second,
	}
	host, err := url.Parse(settings.Fusion.URL)
	if err != nil {
		panic(err)
	}

	return fusionauth.NewClient(client, host, settings.Fusion.APIKey)
}

// new a report service
func newReportService(settings *config.Config, producer *kafka.Producer) *ReportService {
	fusionClient := newFusionClient(settings)

	return &ReportService{
		fusionClient: fusionClient,
		producer:     producer,
	}
}

// validate the token for every requests
func (s *ReportService) validate(token string) (bool, error) {
	result, err := s.fusionClient.ValidateJWT(token)
	if err != nil {
		return false, err
	}
	return result.StatusCode == 200, nil
}

// Report implements gateway.ReportService
func (s *ReportService) Report(ctx context.Context, request *gateway.ReportRequest) (*gateway.ReportReply, error) {
	// validate the token
	ok, err := s.validate(request.Token)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("Authorization failed")
	}

	// send data to kafak
	if err := s.producer.Send(request.Data); err != nil {
		return nil, err
	}

	return &gateway.ReportReply{Status: Success}, nil
}

// start a grpc server
func startServer(addr string, rs *ReportService, wg *sync.WaitGroup) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	gateway.RegisterReportServiceServer(s, rs)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.Serve(lis); err != nil {
			fmt.Printf("grpc serve: %v\n", err)
		}
	}()

	return s, nil
}

func main() {
	var settings config.Config
	// parse the config file
	if err := config.ParseYamlFile("config.yml", &settings); err != nil {
		panic(err)
	}

	// create a kafka producer
	producer := kafka.NewProducer(&settings)

	// new a report service for grpc server
	serivce := newReportService(&settings, producer)

	var wg sync.WaitGroup
	// start a server with listen address
	grpcServer, err := startServer(settings.Server.ListenAddr, serivce, &wg)
	if err != nil {
		panic(err)
	}

	// start the http server
	httpServer := server.NewServer(&settings)
	httpServer.Start(&wg)

	fmt.Println("server is started")

	sig := make(chan os.Signal, 1024)
	// subscribe signals: SIGINT & SINGTERM
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			fmt.Printf("receive signal: %v\n", s)

			// close the grpc server gracefully
			grpcServer.GracefulStop()
			// close the http server gracefully
			httpServer.Stop()

			// wait for server goroutine exit first
			wg.Wait()

			// release the hard resources
			producer.Close()

			fmt.Println("server is shut down")
			return
		}
	}
}
