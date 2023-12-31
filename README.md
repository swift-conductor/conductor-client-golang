# Swift Conductor Client SDK for Go

The `conductor-client-golang` repository provides the client SDKs to build task workers in Go.

Building the task workers in Go mainly consists of the following steps:

1. Setup conductor-client-golang package
2. Create and run task workers
3. Create workflows using code
4. [API Documentation](https://github.com/swift-conductor/conductor-client-golang/tree/main/docs)
   
### Setup Conductor Go Package​

* Create a folder to build your package
```shell
mkdir quickstart/
cd quickstart/
go mod init quickstart
```

* Get Conductor Go SDK

```shell
go get github.com/swift-conductor/conductor-client-golang
```
## Configurations

### Configure API Client
```go

apiClient := client.NewAPIClient(
    settings.NewHttpSettings(
        "http://localhost:8080/api",
    ),
)
	
```

### Setup Logging
SDK uses [logrus](https://github.com/sirupsen/logrus) for logging.

```go
func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}
```

### Next: [Create and run task workers](workers_sdk.md)
