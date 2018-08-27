#!/bin/sh

export CC=i686-w64-mingw32-gcc
#export CFLAGS="-m32"
export CGO_ENABLED=1
export GOARCH=386
export GOOS=windows
go build -o vulkancube.exe main.go
