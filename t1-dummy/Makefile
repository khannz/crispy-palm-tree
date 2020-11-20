export PATH := ./bin:$(PATH)
export GO111MODULE := on

dl-mods:
	go mod download

grpc:
	mkdir -p grpc-dummy
	mkdir -p grpc-orch
	protoc -I ../proto/ --go_out=./grpc-dummy/ --go-grpc_out=./grpc-dummy/ ../proto/dummy.proto
	protoc -I ../proto/ --go_out=./grpc-orch/ --go-grpc_out=./grpc-orch/ ../proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
build-rpm-snapshot:
	goreleaser --snapshot --skip-publish --rm-dist
build-bin:
	CGO_ENABLED=0 go build -o lbost1ad

clean:
	rm -rf dist
	rm -f lbost1ad