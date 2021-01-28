export PATH := ./bin:$(PATH)
export GO111MODULE := on

dummy-dl-mods:
	cd t1-dummy && go mod download

dummy-grpc:
	mkdir -p ./t1-dummy/grpc-dummy
	mkdir -p ./t1-dummy/grpc-orch
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-dummy/ --go-grpc_out=./t1-dummy/grpc-dummy/ ./proto/t1-dummy.proto
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-orch/ --go-grpc_out=./t1-dummy/grpc-orch/ ./proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
dummy-rpm-snapshot:
	cd t1-dummy && goreleaser --snapshot --skip-publish --rm-dist

dummy-bin:
	cd t1-dummy && CGO_ENABLED=0 go build -o lbost1ad

dummy-clean:
	rm -rf ./t1-dummy/dist
	rm -f ./t1-dummy/lbost1ad

ipruler-dl-mods:
	cd t1-ipruler && go mod download

ipruler-grpc:
	mkdir -p ./t1-ipruler/grpc-ipruler
	mkdir -p ./t1-ipruler/grpc-orch
	protoc -I ./proto/ --go_out=./t1-ipruler/grpc-ipruler/ --go-grpc_out=./t1-ipruler/grpc-ipruler/ ./proto/t1-ipruler.proto
	protoc -I ./proto/ --go_out=./t1-ipruler/grpc-orch/ --go-grpc_out=./t1-ipruler/grpc-orch/ ./proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
ipruler-rpm-snapshot:
	cd t1-ipruler && goreleaser --snapshot --skip-publish --rm-dist

ipruler-bin:
	cd t1-ipruler && CGO_ENABLED=0 go build -o lbost1aipr

ipruler-clean:
	rm -rf ./t1-ipruler/dist
	rm -f ./t1-ipruler/lbost1aipr

tunnel-dl-mods:
	cd t1-tunnel && go mod download

tunnel-grpc:
	mkdir -p ./t1-tunnel/grpc-tunnel
	mkdir -p ./t1-tunnel/grpc-orch
	protoc -I ./proto/ --go_out=./t1-tunnel/grpc-tunnel/ --go-grpc_out=./t1-tunnel/grpc-tunnel/ ./proto/t1-tunnel.proto
	protoc -I ./proto/ --go_out=./t1-tunnel/grpc-orch/ --go-grpc_out=./t1-tunnel/grpc-orch/ ./proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
tunnel-rpm-snapshot:
	cd t1-tunnel && goreleaser --snapshot --skip-publish --rm-dist

tunnel-bin:
	cd t1-tunnel && CGO_ENABLED=0 go build -o lbost1at

tunnel-clean:
	rm -rf ./t1-tunnel/dist
	rm -f ./t1-tunnel/lbost1at

route-dl-mods:
	cd t1-route && go mod download

route-grpc:
	mkdir -p ./t1-route/grpc-route
	mkdir -p ./t1-route/grpc-orch
	protoc -I ./proto/ --go_out=./t1-route/grpc-route/ --go-grpc_out=./t1-route/grpc-route/ ./proto/t1-route.proto
	protoc -I ./proto/ --go_out=./t1-route/grpc-orch/ --go-grpc_out=./t1-route/grpc-orch/ ./proto/t1-orch.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
route-rpm-snapshot:
	cd t1-route && goreleaser --snapshot --skip-publish --rm-dist

route-bin:
	cd t1-route && CGO_ENABLED=0 go build -o lbost1ar

route-clean:
	rm -rf ./t1-route/dist
	rm -f ./t1-route/lbost1ad

orch-dl-mods:
	cd t1-orch && go mod download

orch-grpc:
	mkdir -p ./t1-orch/grpc-orch
	protoc -I ./proto/ --go_out=./t1-orch/grpc-orch/ --go-grpc_out=./t1-orch/grpc-orch/ ./proto/t1-orch.proto
	mkdir -p ./t1-orch/grpc-dummy
	protoc -I ./proto/ --go_out=./t1-orch/grpc-dummy/ --go-grpc_out=./t1-orch/grpc-dummy/ ./proto/t1-dummy.proto
	mkdir -p ./t1-orch/grpc-route
	protoc -I ./proto/ --go_out=./t1-orch/grpc-route/ --go-grpc_out=./t1-orch/grpc-route/ ./proto/t1-route.proto
	mkdir -p ./t1-orch/grpc-tunnel
	protoc -I ./proto/ --go_out=./t1-orch/grpc-tunnel/ --go-grpc_out=./t1-orch/grpc-tunnel/ ./proto/t1-tunnel.proto
	mkdir -p ./t1-orch/grpc-ipruler
	protoc -I ./proto/ --go_out=./t1-orch/grpc-ipruler/ --go-grpc_out=./t1-orch/grpc-ipruler/ ./proto/t1-ipruler.proto
	mkdir -p ./t1-orch/grpc-ipvs
	protoc -I ./proto/ --go_out=./t1-orch/grpc-ipvs/ --go-grpc_out=./t1-orch/grpc-ipvs/ ./proto/t1-ipvs.proto
	mkdir -p ./t1-orch/grpc-healthcheck
	protoc -I ./proto/ --go_out=./t1-orch/grpc-healthcheck/ --go-grpc_out=./t1-orch/grpc-healthcheck/ ./proto/t1-healthcheck.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
orch-rpm-snapshot:
	cd t1-orch && goreleaser --snapshot --skip-publish --rm-dist

orch-bin:
	cd t1-orch && CGO_ENABLED=0 go build -o lbost1ao

orch-clean:
	rm -rf ./t1-orch/dist
	rm -f ./t1-orch/lbost1ao

healthcheck-dl-mods:
	cd t1-healthcheck && go mod download

healthcheck-grpc:
	mkdir -p ./t1-healthcheck/grpc-healthcheck
	protoc -I ./proto/ --go_out=./t1-healthcheck/grpc-healthcheck/ --go-grpc_out=./t1-healthcheck/grpc-healthcheck/ ./proto/t1-healthcheck.proto

# TODO: build with flags: go generate & CGO_ENABLED=0 go build -o lbost1ah -ldflags="-X 'github.com/khannz/crispy-palm-tree/cmd.version=v0.2.0' -X 'github.com/khannz/crispy-palm-tree/cmd.buildTime=$(date)'"
healthcheck-rpm-snapshot:
	cd t1-healthcheck && goreleaser --snapshot --skip-publish --rm-dist

healthcheck-bin:
	cd t1-healthcheck && CGO_ENABLED=0 go build -o lbost1ah

healthcheck-clean:
	rm -rf ./t1-healthcheck/dist
	rm -f ./t1-healthcheck/lbost1ah

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

all-grpc:
	mkdir -p ./t1-orch/grpc-orch
	protoc -I ./proto/ --go_out=./t1-orch/grpc-orch/ --go-grpc_out=./t1-orch/grpc-orch/ ./proto/t1-orch.proto
	mkdir -p ./t1-orch/grpc-dummy
	protoc -I ./proto/ --go_out=./t1-orch/grpc-dummy/ --go-grpc_out=./t1-orch/grpc-dummy/ ./proto/t1-dummy.proto
	mkdir -p ./t1-orch/grpc-route
	protoc -I ./proto/ --go_out=./t1-orch/grpc-route/ --go-grpc_out=./t1-orch/grpc-route/ ./proto/t1-route.proto
	mkdir -p ./t1-orch/grpc-tunnel
	protoc -I ./proto/ --go_out=./t1-orch/grpc-tunnel/ --go-grpc_out=./t1-orch/grpc-tunnel/ ./proto/t1-tunnel.proto
	mkdir -p ./t1-orch/grpc-ipruler
	protoc -I ./proto/ --go_out=./t1-orch/grpc-ipruler/ --go-grpc_out=./t1-orch/grpc-ipruler/ ./proto/t1-ipruler.proto
	mkdir -p ./t1-orch/grpc-ipvs
	protoc -I ./proto/ --go_out=./t1-orch/grpc-ipvs/ --go-grpc_out=./t1-orch/grpc-ipvs/ ./proto/t1-ipvs.proto
	mkdir -p ./t1-orch/grpc-healthcheck
	protoc -I ./proto/ --go_out=./t1-orch/grpc-healthcheck/ --go-grpc_out=./t1-orch/grpc-healthcheck/ ./proto/t1-healthcheck.proto
	mkdir -p ./t1-dummy/grpc-dummy
	mkdir -p ./t1-dummy/grpc-orch
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-dummy/ --go-grpc_out=./t1-dummy/grpc-dummy/ ./proto/t1-dummy.proto
	protoc -I ./proto/ --go_out=./t1-dummy/grpc-orch/ --go-grpc_out=./t1-dummy/grpc-orch/ ./proto/t1-orch.proto
	mkdir -p ./t1-ipruler/grpc-ipruler
	mkdir -p ./t1-ipruler/grpc-orch
	protoc -I ./proto/ --go_out=./t1-ipruler/grpc-ipruler/ --go-grpc_out=./t1-ipruler/grpc-ipruler/ ./proto/t1-ipruler.proto
	protoc -I ./proto/ --go_out=./t1-ipruler/grpc-orch/ --go-grpc_out=./t1-ipruler/grpc-orch/ ./proto/t1-orch.proto
	mkdir -p ./t1-tunnel/grpc-tunnel
	mkdir -p ./t1-tunnel/grpc-orch
	protoc -I ./proto/ --go_out=./t1-tunnel/grpc-tunnel/ --go-grpc_out=./t1-tunnel/grpc-tunnel/ ./proto/t1-tunnel.proto
	protoc -I ./proto/ --go_out=./t1-tunnel/grpc-orch/ --go-grpc_out=./t1-tunnel/grpc-orch/ ./proto/t1-orch.proto
	mkdir -p ./t1-route/grpc-route
	mkdir -p ./t1-route/grpc-orch
	protoc -I ./proto/ --go_out=./t1-route/grpc-route/ --go-grpc_out=./t1-route/grpc-route/ ./proto/t1-route.proto
	protoc -I ./proto/ --go_out=./t1-route/grpc-orch/ --go-grpc_out=./t1-route/grpc-orch/ ./proto/t1-orch.proto
	mkdir -p ./t1-healthcheck/grpc-healthcheck
	protoc -I ./proto/ --go_out=./t1-healthcheck/grpc-healthcheck/ --go-grpc_out=./t1-healthcheck/grpc-healthcheck/ ./proto/t1-healthcheck.proto
	mkdir -p ./t1-ipvs/grpc-ipvs
	mkdir -p ./t1-ipvs/grpc-orch
	protoc -I ./proto/ --go_out=./t1-ipvs/grpc-ipvs/ --go-grpc_out=./t1-ipvs/grpc-ipvs/ ./proto/t1-ipvs.proto
	protoc -I ./proto/ --go_out=./t1-ipvs/grpc-orch/ --go-grpc_out=./t1-ipvs/grpc-orch/ ./proto/t1-orch.proto
