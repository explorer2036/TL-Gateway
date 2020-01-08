package main

import (
	"TL-Gateway/config"
	"TL-Gateway/engine"
	"TL-Gateway/kafka"
	"TL-Gateway/proto/gateway"
	"TL-Gateway/report"
	"TL-Gateway/server"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	// DefaultKeepaliveMinTime - if a client pings more than once every 5 seconds, terminate the connection
	DefaultKeepaliveMinTime = 5
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             DefaultKeepaliveMinTime * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,                                  // Allow pings even when there are no active streams
}

// start a grpc server
func startGRPCServer(settings *config.Config, rs *report.Service, wg *sync.WaitGroup) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", settings.Server.ListenAddr)
	if err != nil {
		return nil, err
	}

	// must support the keepalive
	s := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep))
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
	// create a report service for grpc server
	service := report.NewService(&settings)

	// create the engine for handling the messages
	enginer := engine.NewEngine(&settings, producer, service)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	// start the goroutines for sending the messages to kafka
	enginer.Start(ctx, &wg)
	enginer.IsReady()

	// start a server with listen address
	grpcServer, err := startGRPCServer(&settings, service, &wg)
	if err != nil {
		panic(err)
	}

	// start the http server
	httpServer := server.NewServer(&settings)
	httpServer.Start(&wg)

	fmt.Println("gateway is started")

	sig := make(chan os.Signal, 1024)
	// subscribe signals: SIGINT & SINGTERM
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			fmt.Printf("receive signal: %v\n", s)

			start := time.Now()

			// close the grpc server gracefully
			grpcServer.GracefulStop()
			// close the http server gracefully
			httpServer.Stop()

			// cancel the goroutines which is responsible for sending messages to kafka
			cancel()

			// wait for server goroutine exit first
			wg.Wait()

			// release the hard resources
			producer.Close()

			fmt.Printf("shut down takes time: %v\n", time.Now().Sub(start))
			return
		}
	}
}
