# GomeLift

## What's this?

Unofficial GameLift Server SDK in Go.

## How to use

Run `go get` command and import to your program.

```
go get -u github.com/neguse/gomelift

import (
	"github.com/neguse/gomelift/pkg/gamelift"
	glog "github.com/neguse/gomelift/pkg/log"
	"github.com/neguse/gomelift/pkg/proto/pbuffer"
)
```

see [example](example/gamelift/server/server.go).

## How to bulid example.

Build and upload gamelift build.

```
On Linux:
cd /path/to/github.com/neguse/gomelift
go build -o example/gamelift/server/server.exe example/gamelift/server/server.go
aws gamelift upload-build --operating-system AMAZON_LINUX_2 --build-root example/gamelift/server --name gomelift --build-version 1 --region ap-northeast-1
```

If you are working on Windows or OSX, set GOOS=linux and GOARCH=amd64 before you run `go build` and reset it before `go run`.

## License

Copyright 2020 neguse

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
