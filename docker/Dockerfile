#
# The Mega Dockerfile
#
# This dockerfile is an attempt to bundle the following components into 
# one big dockerfile:
#
# - [x] Goevmlab binary 'generic-fuzzer'
# - [x] Go-ethereum binary 'evm'
# - [ ] Erigon binary 'evm'
# - [ ] EvmOne vm binary 'evmone'
# - [x] Reth VM binary 'revme' 
# - [ ] Besu
# - [x] Nethermind
# - [x] Nimbus-eth1
#

# ----------------------------------------------------------
# These builds are performed with golang:latest, which is based on a Debian image
# ----------------------------------------------------------

#
# GETH
#

# golang defaults to a debian flavour)
FROM golang:latest as debian-builder 
RUN git clone https://github.com/ethereum/go-ethereum --depth 1
RUN cd go-ethereum && go run build/ci.go install -static ./cmd/evm && cd ..

#
# Go-evmlab
#

RUN git clone https://github.com/holiman/goevmlab --depth 1
RUN cd goevmlab && go build ./cmd/generic-fuzzer && cd ..

#
# EVMONE 
#

#RUN apt-get update -q && apt-get install -qy --no-install-recommends \
#    ca-certificates g++ cmake ninja-build libgmp-dev

#RUN git clone https://github.com/ethereum/evmone.git --depth 1 --recurse-submodules
#RUN cd evmone && \
#    cmake -S . -B build -DEVMONE_TESTING=ON -DEVMONE_PRECOMPILES_SILKPRE=1
#RUN cd evmone && cmake --build build --parallel
#RUN ls -laR /evmone

#
# NIMBUS-ETH1
#


RUN git clone https://github.com/status-im/nimbus-eth1.git --depth 1 --recurse-submodules
RUN apt update && apt install -y make
RUN cd nimbus-eth1 && make -j4 update 
RUN cd nimbus-eth1 && make -j8 evmstate 
RUN cp nimbus-eth1/tools/evmstate/evmstate /evmstate

# Erigon
RUN git clone https://github.com/ledgerwatch/erigon --depth 1
RUN cd erigon && make evm && cp ./build/bin/evm /erigon_vm



#---------------------------------------------------------------
# Another builder-image (cargo)
#---------------------------------------------------------------

#
# RETH
#

FROM lukemathwalker/cargo-chef:latest-rust-1 AS rust-builder
RUN apt-get update && apt-get -y upgrade && apt-get install -y libclang-dev pkg-config
RUN git clone https://github.com/bluealloy/revm.git --depth 1 
RUN cd revm && cargo build --release --package revme


#---------------------------------------------------------------
# Another builder-image (dotnet)
#---------------------------------------------------------------


#
# NETHERMIND
#

FROM mcr.microsoft.com/dotnet/sdk:8.0-jammy AS dotnet-builder
RUN git clone https://github.com/NethermindEth/nethermind --depth 1 --recurse-submodules
RUN cd nethermind/src/Nethermind/Nethermind.Test.Runner && dotnet publish --self-contained true -r linux-x64 -c Release
RUN mkdir /out && mv nethermind/src/Nethermind/artifacts/bin/Nethermind.Test.Runner/release_linux-x64 /out/neth


#
# Main non-builder
#

FROM debian

#nethtest requires libssl-dev
RUN apt update && apt install -y libssl-dev

COPY --from=debian-builder /go/go-ethereum/build/bin/evm /gethvm
COPY --from=debian-builder /erigon_vm /erigon_vm
COPY --from=debian-builder /go/goevmlab/generic-fuzzer /generic-fuzzer
COPY --from=debian-builder /evmstate /nimbvm
#COPY --from=debian-builder /go/evmone/build/bin/evm /evmone
COPY --from=rust-builder /revm/target/release/revme /revme
COPY --from=dotnet-builder /out/neth /neth
RUN ln -s /neth/nethtest /nethtest

#ENTRYPOINT ["/bin/bash", "-l", "-c"]
ENTRYPOINT ["/generic-fuzzer","--gethbatch=/gethvm","--nethbatch=/nethtest","--nimbus=/nimbvm","--revme=/revme", "--erigonbatch=/erigon_vm", "--fork=Cancun"]
