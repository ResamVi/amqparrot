package main

import (
	"flag"
	"fmt"
	"github.com/resamvi/amqparrot/server"
	"log"
	"os"
	"runtime/debug"
)

func usage() {
	fmt.Fprintf(os.Stdout, `usage: amqparrot [flags]
	-h, --help          show this help
	-v, --version       show version
	-p, --port <PORT>   specify on which port to listen
`)
}

var (
	port        int
	showVersion bool
)

const (
	defaultPort = 8080
)

func main() {
	flag.IntVar(&port, "port", defaultPort, "")
	flag.IntVar(&port, "p", defaultPort, "")
	flag.BoolVar(&showVersion, "version", false, "")
	flag.BoolVar(&showVersion, "v", false, "")

	flag.Usage = usage
	flag.Parse()

	if showVersion {
		info, _ := debug.ReadBuildInfo()
		fmt.Printf("amqparrot %v (%v)\n", info.Deps[0].Version, info.GoVersion)
		return
	}

	srv := server.Server{
		Port: port,
		Log:  log.New(os.Stdout, "", log.LstdFlags),
	}
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
