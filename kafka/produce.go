package kafka

import (
	"TL-Gateway/config"

	"github.com/Shopify/sarama"
)

// Producer for sending data to kafaka
type Producer struct {
	producer sarama.SyncProducer
	settings *config.Config
}

// NewProducer returns a new producer with config
func NewProducer(settings *config.Config) *Producer {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(settings.Kafka.Brokers, config)
	if err != nil {
		panic(err)
	}

	return &Producer{
		producer: producer,
		settings: settings,
	}
}

// Send the bytes to kafka
func (s *Producer) Send(data []byte) error {
	message := &sarama.ProducerMessage{
		Topic:     s.settings.Kafka.Topic,
		Value:     sarama.ByteEncoder(data),
		Partition: int32(-1),
	}

	if _, _, err := s.producer.SendMessage(message); err != nil {
		return err
	}

	return nil
}

// Close release the connection
func (s *Producer) Close() {
	if s.producer != nil {
		s.producer.Close()
	}
}
