#!/usr/bin/env bash
export GOPATH="$(pwd)/vendor"
# Set the GOPROXY environment variable
export GOPROXY=https://goproxy.io
echo ========================
echo Require packge
go mod tidy

echo ========================
echo build darwin_386_ingress
GOOS=darwin GOARCH=386 go build -ldflags "-s -w" -o bin/darwin_386_ingress Ingress.go

echo ========================
echo build darwin_amd64_ingress
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/darwin_amd64_ingress Ingress.go

echo ========================
echo build linux_386_ingress
GOOS=linux GOARCH=386 go build -ldflags "-s -w" -o bin/linux_386_ingress Ingress.go

echo ========================
echo build linux_amd64_ingress
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/linux_amd64_ingress Ingress.go

echo ========================
echo build linux_arm_ingress
GOOS=linux GOARCH=arm go build -ldflags "-s -w" -o bin/linux_arm_ingress Ingress.go

echo ========================
echo build linux_arm64_ingress
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/linux_arm64_ingress Ingress.go

echo ========================
echo build windows_386_ingress.exe
GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o bin/windows_386_ingress.exe Ingress.go

echo ========================
echo build windows_amd64_ingress.exe
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/windows_amd64_ingress.exe Ingress.go

echo ========================
echo Successful
echo ========================
