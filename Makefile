all:
	cp ../TL-Proto/gateway/service.pb.go ./proto/gateway/
	cp ../TL-Proto/id/service.pb.go ./proto/id/
	GO111MODULE=off go build -o TL-Gateway gateway.go
