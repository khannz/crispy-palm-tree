export PATH := ./bin:$(PATH)
export GO111MODULE := on

dl-mods:
	go mod download

grpc:
	mkdir -p grpc-transport
	protoc -I ../proto/ --go_out=./grpc-transport/ --go-grpc_out=./grpc-transport/ ../proto/dummy.proto
	protoc -I ../proto/ --go_out=./grpc-transport/ --go-grpc_out=./grpc-transport/ ../proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
build-rpm-snapshot:
	goreleaser --snapshot --skip-publish --rm-dist
build-bin:
	CGO_ENABLED=0 go build -o lbost1ad

clean:
	rm -rf dist
	rm -rf grpc-transport
	rm -f lbost1ad