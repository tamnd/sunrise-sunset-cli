---
title: "Installation"
description: "Install sunrise-sunset from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/sunrise-sunset-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `sunrise-sunset` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/sunrise-sunset-cli/cmd/sunrise-sunset@latest
```

That puts `sunrise-sunset` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/sunrise-sunset-cli
cd sunrise-sunset-cli
make build        # produces ./bin/sunrise-sunset
./bin/sunrise-sunset version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/sunrise-sunset:latest --help
```

## Checking the install

```bash
sunrise-sunset version
```

prints the version and exits.
