export PATH := ./bin:$(PATH)
export GO111MODULE := on

ipvs-dl-mods:
	cd t1-ipvs && go mod download

ipvs-grpc:
	mkdir -p ./t1-ipvs/grpc-ipvs
	mkdir -p ./t1-ipvs/grpc-orch
	protoc -I ./proto/ --go_out=./t1-ipvs/grpc-ipvs/ --go-grpc_out=./t1-ipvs/grpc-ipvs/ ./proto/t1-ipvs.proto
	protoc -I ./proto/ --go_out=./t1-ipvs/grpc-orch/ --go-grpc_out=./t1-ipvs/grpc-orch/ ./proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
ipvs-rpm-snapshot:
	cd t1-ipvs && goreleaser --snapshot --skip-publish --rm-dist

ipvs-bin:
	cd t1-ipvs && CGO_ENABLED=0 go build -o lbost1ai

ipvs-clean:
	rm -rf ./t1-ipvs/dist
	rm -f ./t1-ipvs/lbost1ai
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
