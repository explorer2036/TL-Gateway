package kafka

import (
	"TL-Gateway/config"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/Shopify/sarama"
)

// Producer for sending data to kafaka
type Producer struct {
	producer sarama.SyncProducer
	settings *config.Config
}

// NewProducer returns a new producer with config
func NewProducer(settings *config.Config) *Producer {
	s := &Producer{settings: settings}

	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	// check if it's TLS connection
	if s.settings.Kafka.Switch {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = s.newTlsConfig()
	}

	// new a synchronizing producer
	producer, err := sarama.NewSyncProducer(settings.Kafka.Brokers, config)
	if err != nil {
		panic(err)
	}
	s.producer = producer

	return s
}

// create tls config with certification files
func (s *Producer) newTlsConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair(s.settings.Kafka.Perm, s.settings.Kafka.Key)
	if err != nil {
		panic(err)
	}

	ca, err := ioutil.ReadFile(s.settings.Kafka.Ca)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic("append certs from pem")
	}

	tlsConfig := &tls.Config{}
	tlsConfig.RootCAs = certPool
	tlsConfig.Certificates = []tls.Certificate{cert}
	tlsConfig.BuildNameToCertificate()

	return tlsConfig
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
