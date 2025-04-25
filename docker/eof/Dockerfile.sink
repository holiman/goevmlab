#
# The EOF Mega Dockerfile
#
# This dockerfile is an attempt to bundle the following components into
# one big dockerfile:
#
# - [ ] Goevmlab binary 'eof-fuzzer'
# - [x] Go-ethereum binary 'evm' (which includes eofparse)
# - [x] Erigon binary 'evm' (which includes eofparse)
# - [x] EvmOne vm binary 'evmone-eofparse'
# - [ ] Reth VM binary 'revme'
# - [x] Besu evmtool (which includes eofparse)
# - [x] Nethermind eofparse
# - [ ] Nimbus-eth1
#


#---------------------------------------------------------------
# golang-builder (debian-based)
#---------------------------------------------------------------
FROM golang:latest as golang-builder

#
# Go-evmlab
#

RUN git clone https://github.com/holiman/goevmlab --branch eofparse --depth 1
RUN cd goevmlab && \
  go build ./cmd/eof-cmp

# GETH
#

RUN git clone https://github.com/ethereum/go-ethereum --depth 1
RUN cd go-ethereum && go run build/ci.go install -static ./cmd/evm

#
# Erigon
#
RUN git clone https://github.com/ledgerwatch/erigon --depth 1
RUN cd erigon && mkdir /erigon/ && make evm && \
	cp ./build/bin/evm /erigon_vm && \
        cp $(bash ./turbo/silkworm/silkworm_lib_path.sh) /libsilkworm_capi.so


#---------------------------------------------------------------
# debian-builder
#---------------------------------------------------------------

#
# EVMONE
#
#
# evmone requires g++ v13, which is _not_ available in debian bookworm (the golang image)
# but it works with debian:testing (at the time of writing this)

FROM debian:testing as debian-builder
RUN apt-get update -q && apt-get install -qy --no-install-recommends git make \
    ca-certificates g++ cmake ninja-build libgmp-dev

RUN git clone https://github.com/ethereum/evmone.git --depth 1 --recurse-submodules
RUN cd evmone && cmake -S . -B build -DEVMONE_TESTING=ON -DEVMONE_PRECOMPILES_SILKPRE=1
RUN cd evmone && cmake --build build --parallel
RUN cp /evmone/build/bin/evmone-eofparse /evmone-eofparse
RUN ls -la /evmone/build/lib/libevmone.so.*
RUN cp /evmone/build/lib/libevmone.so.* /


#---------------------------------------------------------------
# dotnet-builder
#---------------------------------------------------------------


#
# NETHERMIND
#

FROM mcr.microsoft.com/dotnet/sdk:9.0-noble AS dotnet-builder
RUN git clone https://github.com/NethermindEth/nethermind --depth 1 --recurse-submodules


RUN cd nethermind/src/Nethermind/Nethermind.EOFParse.Runner && dotnet publish --self-contained true -r linux-x64 -c Release
RUN mkdir /out && mv nethermind/src/Nethermind/artifacts/bin/Nethermind.EOFParse.Runner/release_linux-x64 /out/neth_eof

#---------------------------------------------------------------
# java-builder
#---------------------------------------------------------------

#
# BESU
#

FROM ubuntu:24.04 as java-builder

RUN apt-get update -q && apt-get install -qy --no-install-recommends git ca-certificates \
  openjdk-21-jdk-headless=21* libjemalloc-dev=5.*
RUN git clone https://github.com/hyperledger/besu.git --depth 1 #--recurse-submodules
RUN cd besu && ./gradlew --parallel ethereum:evmtool:installDist
RUN mkdir /out && mv besu/ethereum/evmtool/build/install/evmtool /out/evmtool

#
# Main non-builder
#

FROM debian:testing

RUN apt-get update -q
# nethtest requires libssl-dev
# besu requires openjdk-21-jre
# evmone requires GLIBC_2.38 (libstdc++-13-dev) (https://github.com/holiman/goevmlab/issues/144)
# EELS requires (see https://github.com/ethereum/execution-specs/issues/976)
#   - pipx, git,
#   - curl (for installing rust),
#   - gcc for building 'py-arkworks-bls12381',
#   - python3-dev for bulding 'lru-dict'
#   - pkg-config (see https://github.com/ethereum/execution-specs/issues/1103)
RUN apt-get install -qy --no-install-recommends \
	libssl-dev  \
	openjdk-21-jre \
	libstdc++-13-dev \
	pipx git curl gcc python3-dev pkg-config \
	jq nano

# Go-evmlab targets
COPY --from=golang-builder /go/goevmlab/eof-cmp /usr/bin
COPY --from=golang-builder /go/goevmlab/cmd/eof-cmp/binaries.txt /binaries.txt

COPY --from=golang-builder /go/go-ethereum/build/bin/evm /gethvm
ENV GETH_BIN=/gethvm

COPY --from=golang-builder /erigon_vm /erigon_vm
COPY --from=golang-builder /libsilkworm_capi.so /lib/libsilkworm_capi.so
ENV ERIG_BIN=/erigon_vm

COPY --from=debian-builder /evmone-eofparse /evmone-eofparse
COPY --from=debian-builder /libevmone.so.* /lib/
ENV EVMO_BIN=/evmone-eofparse

COPY --from=dotnet-builder /out/neth_eof /neth_eof
RUN ln -s /neth/neth_eof /neth_eof
ENV NETH_BIN=/neth/neth_eof

COPY --from=java-builder /out/evmtool /evmtool
RUN ln -s /evmtool/bin/evmtool besu-vm
ENV BESU_BIN=/evmtool/bin/evmtool

COPY readme_docker.md  /README.md


ENTRYPOINT ["/usr/bin/eof-cmp"]
