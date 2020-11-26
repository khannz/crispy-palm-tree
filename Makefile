export PATH := ./bin:$(PATH)
export GO111MODULE := on

dummy-dl-mods:
	cd t1-dummy && go mod download

dummy-grpc:
	mkdir -p ./t1-dummy/grpc-dummy
	mkdir -p ./t1-dummy/grpc-orch
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-dummy/ --go-grpc_out=./t1-dummy/grpc-dummy/ ./proto/dummy.proto
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-orch/ --go-grpc_out=./t1-dummy/grpc-orch/ ./proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
dummy-rpm-snapshot:
	cd t1-dummy && goreleaser --snapshot --skip-publish --rm-dist

dummy-bin:
	cd t1-dummy && CGO_ENABLED=0 go build -o lbost1ad

dummy-clean:
	rm -rf ./t1-dummy/dist
	rm -f ./t1-dummy/lbost1ad

orch-dl-mods:
	cd t1-orch && go mod download

orch-grpc:
	mkdir -p ./t1-orch/grpc-orch
	protoc -I ./proto/ --go_out=./t1-orch/grpc-orch/ --go-grpc_out=./t1-orch/grpc-orch/ ./proto/dummy.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
orch-rpm-snapshot:
	cd t1-orch && goreleaser --snapshot --skip-publish --rm-dist

orch-bin:
	cd t1-orch && CGO_ENABLED=0 go build -o lbost1ao

orch-clean:
	rm -rf ./t1-orch/dist
	rm -f ./t1-orch/lbost1ao