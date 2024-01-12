# Swift Conductor Client SDK for Go

### Install Swift Conductor Client Go Package​

```sh
go get github.com/swift-conductor/conductor-client-golang
```

## Create and start Workflows

[Create and Execute Workflows](docs/readme/workflows.md)

## Create and run task workers 

[Create and run task workers](docs/readme/workers.md)

## API Documentation

[API Documentation](https://github.com/swift-conductor/conductor-client-golang/tree/main/docs/api)

### Setup Logging

SDK uses [logrus](https://github.com/sirupsen/logrus) for logging.

```go
func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}
```
