# Binaries

This folder contains binaries for the various evms that goevmlab can use. 
It will not be updated frequently, as these are pretty huge, but are needed
in order to run some of the testcases. 

They have been compressed in order to not make github choke on them:
```
zip -9 evm  ./evm && zip -9 nethtest  ./nethtest && zip -9 parity-evm  ./parity-evm
```
I had to leave `testeth` out of it, because it's 400+ Mb. The testcases
will decompress `foo.zip` as `foo`, _unless_ `foo` already exists. Therefore, 
you can replace the evm binaries  with more up-to-date version, and run the tests
to ensure that they are still working properly. 

## Build instructions

Most up-to-date build instructions can be found on the various project's pages, 
but, fwiw, here's some info if you want to build locally. 


### Geth (`evm`)

Geth is simple to build, if you have to go toolchain installed. 
```
[user@work go-ethereum]$ make all
```
The command above would generate `build/bin/evm`. It's also possible to
do `go build ./cmd/evm`. 

### Nethermind (`nethtest`)

For nethermind, I have a debian VM that I build on. Here are two scripts for it: 
`install_dotnet.sh`:

```
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > microsoft.asc.gpg
sudo mv microsoft.asc.gpg /etc/apt/trusted.gpg.d/
wget -q https://packages.microsoft.com/config/debian/9/prod.list
sudo mv prod.list /etc/apt/sources.list.d/microsoft-prod.list
sudo chown root:root /etc/apt/trusted.gpg.d/microsoft.asc.gpg
sudo chown root:root /etc/apt/sources.list.d/microsoft-prod.list

sudo apt-get -y install apt-transport-https
sudo apt-get update
# These need to be updated from time to time
#sudo apt-get install dotnet-sdk-2.2
#sudo apt-get -y install dotnet-sdk-3.0
sudo apt-get -y install dotnet-sdk-3.1
```
And `build_nethermind.sh`:
```
sudo apt-get update && sudo apt-get install libsnappy-dev libc6-dev libc6
( cd nethermind/src/Nethermind/Nethermind.State.Test.Runner && \
  dotnet publish -r linux-x64 -c Release /p:PublishSingleFile=true && \
  cp bin/Release/netcoreapp3.1/linux-x64/publish/nethtest ../../../)
```

## Parity (`parity-evm`)

Assuming you have a rust toolchain, then: 

```
cargo build --release -p evmbin
```
Should spit out the binary into `target/release/parity-evm`


## Aleth (`testeth`)

TODO
