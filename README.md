# GomeLift

## What's this?

Unofficial GameLift Server SDK in Go.

## Usage

Run go get command.

```
go get -u github.com/neguse/gomelift
```

Build and upload gamelift build.

```
cd /path/to/github.com/neguse/gomelift
go build -o example/gamelift/server/server.exe example/gamelift/server/server.go
aws gamelift aws gamelift upload-build --operating-system [os] --build-root example/gamelift/server ...
```

see [example/gamelift/server/server.go](example).

## License

TDB
