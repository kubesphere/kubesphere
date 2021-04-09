# build mulit-arch image with docker buildx

## set up env for docker buildx

run follow commands

```
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
docker buildx rm arm64builder
docker buildx create --name arm64builder
docker buildx use arm64builder
docker buildx inspect --bootstrap
```

## build images

If build multi platform one time, can not use `--load` param, You must have privileges to push images to `$REPO` 

```
docker buildx build --platform="linux/amd64,linux/arm64" . -t $REPO/ks-apiserver:$TAG -f build-multiarch/Dockerfile --target=ks-apiserver --push
docker buildx build --platform="linux/amd64,linux/arm64" . -t $REPO/ks-controller-manager:$TAG -f build-multiarch/Dockerfile --target=ks-controller-manager --push
```

Also you can build a image for local build with only one arch

```
docker buildx build --platform="linux/arm64" . -t "ks-apiserver:arm64" -f build-multiarch/Dockerfile --target=ks-apiserver --load
docker buildx build --platform="linux/amd64" . -t "ks-apiserver:amd64" -f build-multiarch/Dockerfile --target=ks-apiserver --load
```