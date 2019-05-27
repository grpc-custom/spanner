proto-fmt:
	find . -name "*.proto" | xargs clang-format -i

protoc:
	protoc -I. \
		-I=${GOPATH}/src \
		--go_out=plugins=grpc:${GOPATH}/src \
		ddl.proto

build-spanner-ddl-debug:
	go build -o protoc-gen-spanner-ddl-debug ./protoc-gen-spanner-ddl/main.go

proto-spanner-ddl-debug: build-spanner-ddl-debug
	protoc \
		-I=${GOPATH}/src:. \
		-I=${GOPATH}/src/github.com/googleapis/googleapis:. \
		-I=${GOPATH}/src/github.com/grpc-custom:. \
		--plugin=./protoc-gen-spanner-ddl-debug \
		--spanner-ddl-debug_out=logtostderr=true:. \
		example/*.proto
