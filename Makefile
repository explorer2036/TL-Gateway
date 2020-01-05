all:
	cp ../TL-Proto/gateway/report.pb.go ./proto/gateway/
	GO111MODULE=off go build -o TL-Gateway gateway.go
