<img src="https://i.imgur.com/NcAPrBo.png" width="125px" align="left">

<h3>Mimicks a RabbitMQ server and prints any events it receives to stdout</h3>
<div>
<p>
Is a bare-bones AMQP server implementation working as a drop-in replacement that receives and acknowledges any messages it was sent. 

<hr>

![anim](https://user-images.githubusercontent.com/6261556/167254191-aa61f696-47b8-4e0c-9a77-c5e514f82207.gif)

[![Go Report Card](https://goreportcard.com/badge/github.com/resamvi/amqparrot)](https://goreportcard.com/report/github.com/resamvi/amqparrot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/ResamVi/amqparrot/blob/master/LICENSE)

	
## Installation
Install via
```
go install github.com/resamvi/amqparrot@latest
```

Or download a pre-compiled binary from here

| Version 1.0.0    |                                                                                                          |
|------------------|----------------------------------------------------------------------------------------------------------|
| Windows (64-bit) | <a href="https://github.com/ResamVi/spayle/releases/download/1.0.0-alpha/windows-x32.zip">Download</a><br> |
| Mac              | <a href="https://github.com/ResamVi/spayle/releases/download/1.0.0-alpha/mac-x64.zip">Download</a>       |
| Linux            | <a href="https://github.com/ResamVi/spayle/releases/download/1.0.0-alpha/linux-x32.tar">Download</a>     |
| Source           | <a href="https://github.com/ResamVi/spayle/releases/download/1.0.0-alpha/linux-x32.tar">Download</a>     |


## Usage

Start amqparrot via
```
amqparrot
```
	
Or if running locally navigate
```
go run amqparrot.go
```

Start an example client that publishes events (see `example/example.go`)
```go
package main

import (
	"fmt"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:8080//dev")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Created connection")

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	fmt.Println("Created channel")

	err = ch.Publish("example-exchange", "my.routing.key", true, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte("Hello World"),
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("Published message")

	ch.Close()
	fmt.Println("closed channel")
}

```

Watch the response of amqparrot
```
2022/05/07 14:19:45 Listening on port 8080
2022/05/07 14:19:49 Serving [::1]:56887
2022/05/07 14:19:49 hello received
2022/05/07 14:19:49 connection start ok with user "guest" and pass "guest" using mechanism "PLAIN"
2022/05/07 14:19:49 connection tune ok
2022/05/07 14:19:49 Connection created in vhost '/dev'
2022/05/07 14:19:49 Opened a Channel with id 1
2022/05/07 14:19:49 Message to exchange 'example-exchange' with routing key 'my.routing.key'
2022/05/07 14:19:49 Received body:
Hello World
2022/05/07 14:19:49 Closed a Channel with id 1
```

## Details
```
$ amqparrot --help
usage: amqparrot [flags]
        -h, --help          show this help
        -v, --version       show version
        -p, --port <PORT>   specify on which port to listen
```

