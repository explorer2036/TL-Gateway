all:
	cp ../TL-Proto/gateway/report.pb.go ./proto/gateway/
	GO111MODULE=off go build -o gateway gateway.go
