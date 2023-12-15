#!/usr/bin/env sh

for os in windows darwin linux; do
  ext=''
  if [ "$os" = "windows" ]; then
    ext='.exe'
  fi
  for arch in amd64 arm64; do
    CGO_ENABLED=0 GOOS=$os GOARCH=amd64 go build -v -o s5s-$os-$arch$ext -trimpath -ldflags "-s -w -buildid="
  done
done
