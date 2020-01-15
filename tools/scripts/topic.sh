./kafka-topics.sh --zookeeper 192.168.0.6:2181 --create --topic TL-Trader --partitions 30  --replication-factor 1

./kafka-console-consumer.sh --bootstrap-server 192.168.0.4:9092 --topic TL-Trader --from-beginning