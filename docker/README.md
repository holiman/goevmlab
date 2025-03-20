# Dockerfile for all clients

This is a docker container contining all clients and all goevmlab executables


## Build problems

Docker nowadays is aggressively parallel, which may cause problems when building humongous 
clients. To fix that, see [resource limiting](https://docs.docker.com/build/buildkit/configure/#resource-limiting):

```
# /etc/buildkitd.toml
[worker.oci]
  max-parallelism = 4
```
```
 docker buildx create --use \
  --name mybuilder \
  --driver docker-container \
  --config /etc/buildkitd.toml
```
and then build
```
docker buildx   build   --progress=plain -t holiman/omnifuzz .
```
or
```
docker buildx build --progress=plain --push --load  -t holiman/omnifuzz-l2  -f Dockerfile.L2.txt .
```

## The container itself

There's more information in the [in-docker-readme](readme_docker.md)
