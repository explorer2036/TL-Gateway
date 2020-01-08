package report

import (
	"TL-Gateway/config"
	"TL-Gateway/model"
	"TL-Gateway/proto/gateway"
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
)

const (
	// DefaultFustionTimeout defines the default timeout for http request every time
	DefaultFustionTimeout = 5
	// DefaultStreamBuffer - queue size for goroutines which is responsible for sending kafka(default 100)
	DefaultStreamBuffer = 10240

	// Success for grpc requests
	Success = 0
	// ErrorAuthorization for token validation
	ErrorAuthorization = 1
	// ErrorKafka for writing buffer queue
	ErrorKafka = 2
)

// Service represents the interfaces for gateway
type Service struct {
	gateway.UnimplementedReportServiceServer

	fusionClient *fusionauth.FusionAuthClient // fusion client for validate the token
	settings     *config.Config               // settings for the gateway
	stream       chan model.Carrier           // stream for buffering messages
}

// NewService create the report service
func NewService(settings *config.Config) *Service {
	if settings.Cache.Buffer == 0 {
		settings.Cache.Buffer = DefaultStreamBuffer
	}

	s := &Service{settings: settings}

	// init the fusion client
	s.fusionClient = s.newFusionClient()
	// init the buffer queue
	s.stream = make(chan model.Carrier, s.settings.Cache.Buffer)

	return s
}

// new a fusion client
func (s *Service) newFusionClient() *fusionauth.FusionAuthClient {
	client := &http.Client{
		Timeout: s.settings.Fusion.Timeout * time.Second,
	}

	// parse the host from url
	host, err := url.Parse(s.settings.Fusion.URL)
	if err != nil {
		panic(err)
	}

	return fusionauth.NewClient(client, host, s.settings.Fusion.APIKey)
}

// validate the token for every requests
func (s *Service) validate(token string) (bool, error) {
	result, err := s.fusionClient.ValidateJWT(token)
	if err != nil {
		return false, err
	}
	return result.StatusCode == 200, nil
}

// Report implements gateway.ReportService
func (s *Service) Report(ctx context.Context, request *gateway.ReportRequest) (*gateway.ReportReply, error) {
	// validate the token
	// ok, err := s.validate(request.Token)
	// if err != nil {
	// 	return nil, err
	// }
	// if !ok {
	// 	return &gateway.ReportReply{Status: ErrorAuthorization}, nil
	// }

	// send data to buffer queue
	select {
	case s.stream <- request.Data:
	default:
		return &gateway.ReportReply{Status: ErrorKafka}, nil
	}

	return &gateway.ReportReply{Status: Success}, nil
}

// ReadMessages returns the messages from channel
func (s *Service) ReadMessages() <-chan model.Carrier {
	return s.stream
}
