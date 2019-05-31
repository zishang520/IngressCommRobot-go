# Ingress Comm Automatically send message

### Tips:

```
go mod vendor && go build Ingress.go
```
If you are a windows user, you need to have Visual Studio and MinGW installed

## Windows
```cmd
build.bat
```
## Other
```shell
build.sh
```
---------------------------------------

# Run Bin

Linux:

```sh
$ chmod +x Ingress & ./Ingress -time=5 -sleep=120
```

Windows:

```bat
.\Ingress.exe -time=5 -sleep=120
```

---------------------------------------
### Configuration Information:

service/data/conf.json.default modify the configuration and renamed conf.json

--/
---Ingress
---data/
--------conf.json
--------agent.db