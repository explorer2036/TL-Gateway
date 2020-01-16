all:
	cp ../TL-Proto/gateway/service.pb.go ./proto/gateway/
	GO111MODULE=off go build -o TL-Gateway gateway.go
