@echo OFF
set GOPATH="%~dp0vendor"
rem Set the GOPROXY environment variable
set GOPROXY=https://goproxy.io

echo ========================
echo Require packge
go mod tidy

echo ========================
echo build darwin_386_ingress
set GOOS=darwin
set GOARCH=386
go build -ldflags "-s -w" -o bin/darwin_386_ingress Ingress.go

echo ========================
echo build darwin_amd64_ingress
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w" -o bin/darwin_amd64_ingress Ingress.go

echo ========================
echo build linux_386_ingress
set GOOS=linux
set GOARCH=386
go build -ldflags "-s -w" -o bin/linux_386_ingress Ingress.go

echo ========================
echo build linux_amd64_ingress
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o bin/linux_amd64_ingress Ingress.go

echo ========================
echo build linux_arm_ingress
set GOOS=linux
set GOARCH=arm
go build -ldflags "-s -w" -o bin/linux_arm_ingress Ingress.go

echo ========================
echo build linux_arm64_ingress
set GOOS=linux
set GOARCH=arm64
go build -ldflags "-s -w" -o bin/linux_arm64_ingress Ingress.go

echo ========================
echo build windows_386_ingress.exe
set GOOS=windows
set GOARCH=386
go build -ldflags "-s -w" -o bin/windows_386_ingress.exe Ingress.go

echo ========================
echo build windows_amd64_ingress.exe
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o bin/windows_amd64_ingress.exe Ingress.go

echo ========================
echo Successful
echo ========================
pause
