package engine

import (
	"TL-Gateway/config"
	"TL-Gateway/kafka"
	"TL-Gateway/report"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalConsumedCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_consumed_count",
		Help: "The total count of messages consumed to kafka",
	})

	totalSendFailedCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_send_failed_count",
		Help: "The total count of messages sending failed to kafka",
	})
)

func init() {
	prometheus.MustRegister(totalConsumedCount)
}

const (
	// DefaultNumberOfRoutines - number of the goroutines, for sending messages to kafka
	DefaultNumberOfRoutines = 2
	// DefaultDelayTime - waiting for reading channel when receive exit signals(default 2s)
	DefaultDelayTime = 2
	// DefaultRetryTime - retry time seconds when sending messages to kafka happens error
	DefaultRetryTime = 2
)

// Engine is used to implement gateway.Engine.
type Engine struct {
	producer *kafka.Producer // kafka producer for writing messages
	service  *report.Service // report service with stream channel

	settings *config.Config // settings for the gateway
	ready    chan struct{}  // mark the goroutines are ready
}

// NewEngine create a enginer for handling messages
func NewEngine(settings *config.Config, producer *kafka.Producer, service *report.Service) *Engine {
	if settings.Cache.Routines == 0 {
		settings.Cache.Routines = DefaultNumberOfRoutines
	}
	if settings.Cache.DelayTime == 0 {
		settings.Cache.DelayTime = DefaultDelayTime
	}

	return &Engine{
		settings: settings,
		producer: producer,
		service:  service,
		ready:    make(chan struct{}, settings.Cache.Routines),
	}
}

// handle the left data in stream
func (e *Engine) afterCare() bool {
	delay := e.settings.Cache.DelayTime

	select {
	case d := <-e.service.ReadMessages():
		// send the message to kafka
		e.send(d)

	case <-time.After(time.Second * delay):
		return false
	}

	return true
}

// send the message to kafka
func (e *Engine) send(data []byte) {
	// if failed, try to send periodly
	for {
		// try to send data to kafka
		if err := e.producer.Send(data); err != nil {
			fmt.Printf("send messages: %v\n", err)

			// update the metrics
			totalSendFailedCount.Add(1)

			time.Sleep(DefaultRetryTime * time.Second)
			continue
		}

		// update the consumed metric
		totalConsumedCount.Add(1)

		break
	}
}

func (e *Engine) handle(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// mark the goroutine started
	e.ready <- struct{}{}
	for {
		// consider flush the data in channel before exiting

		// try to read the message first
		select {
		case d := <-e.service.ReadMessages():
			// send the message to kafka
			e.send(d)

		case <-ctx.Done():
			// if the cancel signal is received
			// try to handle the left data
			for {
				if e.afterCare() == false {
					fmt.Println("engine goroutine exit")
					return
				}
			}
		}
	}
}

// Start create goroutines to read messages and send to kafka
func (e *Engine) Start(ctx context.Context, wg *sync.WaitGroup) {
	for i := 0; i < e.settings.Cache.Routines; i++ {
		wg.Add(1)
		go e.handle(ctx, wg)
	}
}

// IsReady check if the goroutines is ready
func (e *Engine) IsReady() {
	for i := 0; i < e.settings.Cache.Routines; i++ {
		<-e.ready
	}
	fmt.Println("engine goroutines are ready")
}
