package main

import (
	"TL-Gateway/config"
	"TL-Gateway/engine"
	"TL-Gateway/kafka"
	"TL-Gateway/log"
	"TL-Gateway/proto/gateway"
	"TL-Gateway/report"
	"TL-Gateway/server"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc/credentials"

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

// prepare the files for certification
func createCredentials(settings *config.Config) credentials.TransportCredentials {
	cert, err := tls.LoadX509KeyPair(settings.Server.PermFile, settings.Server.KeyFile)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(settings.Server.CaFile)
	if err != nil {
		panic(err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic("append certs from pem")
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	})
}

// start a grpc server
func startGRPCServer(settings *config.Config, rs *report.Service, wg *sync.WaitGroup) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", settings.Server.ListenAddr)
	if err != nil {
		return nil, err
	}

	var s *grpc.Server
	// must support the keepalive
	if settings.Server.TLSSwitch {
		// create credentials by loading the cert files
		creds := createCredentials(settings)
		s = grpc.NewServer(grpc.Creds(creds), grpc.KeepaliveEnforcementPolicy(kaep))
	} else {
		s = grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep))
	}
	gateway.RegisterReportServiceServer(s, rs)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.Serve(lis); err != nil {
			log.Infof("grpc serve: %v\n", err)
		}
	}()

	return s, nil
}

// updateOptions updates the log options
func updateOptions(scope string, options *log.Options, settings *config.Config) error {
	options.RotateOutputPath = settings.Log.RotationPath
	options.RotationMaxBackups = settings.Log.RotationMaxBackups
	options.RotationMaxSize = settings.Log.RotationMaxSize
	options.RotationMaxAge = settings.Log.RotationMaxAge
	options.JSONEncoding = settings.Log.JSONEncoding
	level, err := options.ConvertLevel(settings.Log.OutputLevel)
	if err != nil {
		return err
	}
	options.SetOutputLevel(scope, level)
	options.SetLogCallers(scope, true)

	return nil
}

func main() {
	var settings config.Config
	// parse the config file
	if err := config.ParseYamlFile("config.yml", &settings); err != nil {
		panic(err)
	}

	// init and update the log options
	logOptions := log.DefaultOptions()
	if err := updateOptions("default", logOptions, &settings); err != nil {
		panic(err)
	}
	// configure the log options
	if err := log.Configure(logOptions); err != nil {
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

	log.Info("gateway is started")

	sig := make(chan os.Signal, 1024)
	// subscribe signals: SIGINT & SINGTERM
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			log.Infof("receive signal: %v", s)

			// flush the log
			log.Sync()

			start := time.Now()

			// close the grpc server gracefully
			grpcServer.GracefulStop()
			// close the http server gracefully
			httpServer.Stop()

			log.Info("server is stopped")

			// cancel the goroutines which is responsible for sending messages to kafka
			cancel()

			// wait for server goroutine exit first
			wg.Wait()

			// release the hard resources
			producer.Close()

			log.Infof("shut down takes time: %v", time.Now().Sub(start))
			return
		}
	}
}
