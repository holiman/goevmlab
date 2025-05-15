# Dockerfile for all clients

This is a docker container contining all clients and all goevmlab executables


## Build problems

Docker nowadays is aggressively parallel, which may cause problems when building humongous 
clients. To fix that, see [resource limiting](https://docs.docker.com/build/buildkit/configure/#resource-limiting):

```
# /etc/buildkitd.toml
[worker.oci]
  max-parallelism = 1
```
```
 docker buildx create --use \
  --name mybuilder \
  --driver docker-container \
  --config /etc/buildkitd.toml
```

## Building

Building the main `holiman/omnifuzz` image:
```
docker buildx build --progress=plain --load --push -t holiman/omnifuzz .
```

However, the `nimbus-eth1` client takes so long to build, so it's been moved into a separate container: `holiman/nimbus`:
```
docker buildx build --progress=plain --load  --push -t holiman/nimbus -f Dockerfile.nimbus  .
```
In order to update _everything_, the `holiman/nimbus`-image must first be rebuilt, then the regular `holiman/omnifuzz`-client needs
to be rebuilt. And the caches needs clearing: 
```
docker system prune -af && docker buildx prune -f
docker buildx build --progress=plain --load  --push -t holiman/nimbus -f Dockerfile.nimbus  .
curl -d "Docker build #1 done"  ntfy.sh/yourchannel
docker buildx build --progress=plain --push --load  -t holiman/omnifuzz  .
curl -d "Docker build #2 done"  ntfy.sh/yourchannel
```

## Running the container

There's more information in the [in-docker-readme](readme_docker.md)
