package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
)

// groupHandler represents a Sarama consumer group consumer
type groupHandler struct {
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (g *groupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (g *groupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

var (
	brokers = flag.String("a", "192.168.0.4:9092", "the grpc server address")
)

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (g *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for {
		select {
		case m, ok := <-claim.Messages():
			// if the messages channel is closed, exit the goroutine
			if !ok {
				return nil
			}

			// send the message to channel
			fmt.Printf("topic: %v, value: %v\n", m.Topic, string(m.Value))

			// mark the message consumed
			session.MarkMessage(m, "")
		}
	}
}

func main() {
	// new a sarama config for consumer group
	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group := "TL-Trader-Group"
	topics := []string{"TL-Trader"}

	// setup a new sarama consumer group
	handler := groupHandler{}

	// new a consumer group client for consuming messages with group
	client, err := sarama.NewConsumerGroup([]string{*brokers}, group, config)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()
	// loop for consuming messages with topics
	for {
		// try to fetch messages from kafka
		if err := client.Consume(ctx, topics, &handler); err != nil {
			fmt.Printf("restart consumer topics(%v), group(%s): %v\n", topics, group, err)
			time.Sleep(time.Second)
			continue
		}

		select {
		// receive a canceled signal
		case <-ctx.Done():
			return
		default:
		}
	}
}
