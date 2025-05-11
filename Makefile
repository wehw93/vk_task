.PHONY: runApp

runApp:
	go run cmd/subscribtion/main.go 

.PHONY: proto

proto:
	protoc --go_out=. --go-grpc_out=. proto/pubsubservice.proto
	
.PHONY: runTestsSubPub

runTestsSubPub:
	go test -timeout 30s -coverprofile=/tmp/vscode-go8S4Khf/go-code-cover vk_task/pkg/subpub 


.DEFAULT_GOAL := runApp