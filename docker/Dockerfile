#
# The Mega Dockerfile
#
# This dockerfile is an attempt to bundle the following components into 
# one big dockerfile:
#
# - [x] Goevmlab binary 'generic-fuzzer'
# - [x] Go-ethereum binary 'evm'
# - [x] Erigon binary 'evm'
# - [x] EvmOne vm binary 'evmone'
# - [x] Reth VM binary 'revme' 
# - [x] Besu
# - [x] Nethermind
# - [x] Nimbus-eth1
#

#---------------------------------------------------------------
# golang-builder (debian-based)
#---------------------------------------------------------------
FROM golang:latest as golang-builder 

#
# Go-evmlab
#

RUN git clone https://github.com/holiman/goevmlab --depth 1
RUN cd goevmlab && \
  go build ./cmd/generic-fuzzer && \
  go build ./cmd/checkslow && \
  go build ./cmd/minimizer && \
  go build ./cmd/repro && \
  go build ./cmd/runtest && \
  go build ./cmd/tracediff && \
  go build ./cmd/traceview

#
# GETH
#

RUN git clone https://github.com/ethereum/go-ethereum --depth 1
RUN cd go-ethereum && go run build/ci.go install -static ./cmd/evm

#
# Erigon
#
RUN git clone https://github.com/ledgerwatch/erigon --depth 1
RUN cd erigon && make evm && cp ./build/bin/evm /erigon_vm
#RUN ls -la /go/pkg/mod/github.com/erigontech/


#
# NIMBUS-ETH1
#

RUN git clone https://github.com/status-im/nimbus-eth1.git --depth 1 --recurse-submodules
RUN apt-get update -q && apt-get install -qy --no-install-recommends make
RUN cd nimbus-eth1 && make -j8 update 
RUN cd nimbus-eth1 && make -j8 evmstate 
RUN cp nimbus-eth1/tools/evmstate/evmstate /evmstate





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
RUN cp /evmone/build/bin/evmone-statetest /evmone-statetest
RUN cp /evmone/build/lib/libevmone.so.0.12 /libevmone.so.0.12
#---------------------------------------------------------------
# rust-builder
#---------------------------------------------------------------

#
# RETH
#

FROM lukemathwalker/cargo-chef:latest-rust-1 AS rust-builder
RUN apt-get update -q && apt-get install -qy --no-install-recommends libclang-dev pkg-config
RUN git clone https://github.com/bluealloy/revm.git --depth 1 
RUN cd revm && cargo build --release --package revme


#---------------------------------------------------------------
# dotnet-builder
#---------------------------------------------------------------


#
# NETHERMIND
#

FROM mcr.microsoft.com/dotnet/sdk:8.0-jammy AS dotnet-builder
RUN git clone https://github.com/NethermindEth/nethermind --depth 1 --recurse-submodules

RUN cd nethermind/src/Nethermind/Nethermind.Test.Runner && dotnet publish --self-contained true -r linux-x64 -c Release
RUN mkdir /out && mv nethermind/src/Nethermind/artifacts/bin/Nethermind.Test.Runner/release_linux-x64 /out/neth

#---------------------------------------------------------------
# java-builder
#---------------------------------------------------------------

#
# BESU 
#

FROM ubuntu:23.10 as java-builder

RUN apt-get update -q && apt-get install -qy --no-install-recommends git ca-certificates 
RUN git clone https://github.com/hyperledger/besu.git --depth 1 #--recurse-submodules
RUN apt-get install -qy --no-install-recommends git openjdk-17-jre-headless=17* libjemalloc-dev=5.* 
RUN cd besu && ./gradlew --parallel ethereum:evmtool:installDist
RUN mkdir /out && mv besu/ethereum/evmtool/build/install/evmtool /out/evmtool

#
# Main non-builder
#

FROM debian:testing

RUN apt-get update -q
# nethtest requires libssl-dev
RUN apt-get install -qy --no-install-recommends libssl-dev
# besu requires openjdk-17-jre
RUN apt-get install -qy --no-install-recommends  openjdk-17-jre 

# Install execution-specs (EELS)
RUN apt-get install -qy --no-install-recommends pipx git && \
  git clone https://github.com/ethereum/execution-specs.git --branch statetests --depth 1 && \
  PIPX_HOME=/opt/pipx PIPX_BIN_DIR=/ pipx install './execution-specs/[test]'
ENV EELS_BIN=/ethereum-spec-evm

# Go-evmlab targets
COPY --from=golang-builder /go/goevmlab/generic-fuzzer /usr/bin
COPY --from=golang-builder /go/goevmlab/checkslow  /usr/bin
COPY --from=golang-builder /go/goevmlab/minimizer /usr/bin
COPY --from=golang-builder /go/goevmlab/repro /usr/bin
COPY --from=golang-builder /go/goevmlab/runtest /usr/bin
COPY --from=golang-builder /go/goevmlab/tracediff /usr/bin
COPY --from=golang-builder /go/goevmlab/traceview /usr/bin

COPY --from=golang-builder /go/go-ethereum/build/bin/evm /gethvm
ENV GETH_BIN=/gethvm

COPY --from=golang-builder /erigon_vm /erigon_vm
COPY --from=golang-builder /go/pkg/mod/github.com/erigontech/silkworm-go@v0.14.0/lib/linux_x64/libsilkworm_capi.so /lib/libsilkworm_capi.so
ENV ERIG_BIN=/erigon_vm

COPY --from=golang-builder /evmstate /nimbvm
ENV NIMB_BIN=/nimbvm

COPY --from=debian-builder /evmone-statetest /evmone
COPY --from=debian-builder /libevmone.so.0.12 /lib/libevmone.so.0.12
ENV EVMO_BIN=/evmone

COPY --from=rust-builder /revm/target/release/revme /revme
ENV RETH_BIN=/revme

COPY --from=dotnet-builder /out/neth /neth
RUN ln -s /neth/nethtest /nethtest
ENV NETH_BIN=/neth/nethtest

COPY --from=java-builder /out/evmtool /evmtool
RUN ln -s /evmtool/bin/evm besu-vm
ENV BESU_BIN=/evmtool/bin/evm

COPY readme_docker.md  /README.md
COPY entrypoint.sh  /entrypoint.sh

ENV FUZZ_CLIENTS="--gethbatch=$GETH_BIN \
 --nethbatch=$NETH_BIN \
 --nimbus=$NIMB_BIN \
 --revme=$RETH_BIN \
 --erigonbatch=$ERIG_BIN \
 --besubatch=$BESU_BIN \
 --evmone=$EVMO_BIN \
 --eelsbatch=$EELS_BIN"

ENTRYPOINT ["/bin/bash"]
