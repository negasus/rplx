.PHONY: proto

proto:
	protoc -I=. --go_out=plugins=grpc:. ./message.proto

mockgen:
	mockgen -source=message.pb.go -destination ./test/mock.go -package test
