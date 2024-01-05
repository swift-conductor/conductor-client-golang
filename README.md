# Swift Conductor Client SDK for Go

### Install Swift Conductor Client Go Package​

```sh
go get swiftconductor.com/swift-conductor-client
```

## Configuration

### API Client

```go
apiClient := client.NewAPIClient(settings.NewHttpSettings("http://localhost:8080/api",))
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

## Create and run task workers 

[Create and run task workers](docs/readme/workers.md)

## Create and Execute Workflows

[Create and Execute Workflows](docs/readme/workflows.md)

## API Documentation

[API Documentation](https://github.com/swift-conductor/conductor-client-golang/tree/main/docs/api)