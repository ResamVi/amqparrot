package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"

	"github.com/resamvi/amqparrot/amqp"
	"github.com/resamvi/amqparrot/amqp/client"
	"github.com/resamvi/amqparrot/amqp/server"
)

// Logger allows for inserting a custom logger with custom format
type Logger interface {
	Printf(format string, v ...any)
}

type Server struct {
	// Port to listen for tcp connections.
	Port int

	// Log defines how the server prints its log messages
	// Default: log.New(os.Stdout, "", log.LstdFlags)
	Log Logger
}

// Start the server
func (s Server) Start() error {
	if s.Log == nil {
		s.Log = log.New(os.Stdout, "", 0)
	}

	lstner, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return fmt.Errorf("could not start server: %w", err)
	}
	s.Log.Printf(Started, s.Port)

	for {
		conn, err := lstner.Accept()
		if err != nil {
			s.Log.Printf("Error accepting TCP connection: %v", err)
			continue
		}
		s.Log.Printf("Serving %s\n", conn.RemoteAddr().String())

		msgStream := Stream(conn)
		go func() {
			for msg := range msgStream {
				s.handle(msg, conn)
			}
		}()
	}
}

const (
	lineEscape = "\n"

	// All possible log lines server sends out

	Started = "Listening on port %v"

	Hello = "hello received" + lineEscape

	ConnectionStartOk = "connection start ok with user \"%s\" and pass \"%s\" using mechanism \"%s\"" + lineEscape
	ConnectionTuneOk  = "connection tune ok" + lineEscape
	ConnectionOpen    = "Connection created in vhost '%v'" + lineEscape

	ChannelOpen  = "Opened a Channel with id %v" + lineEscape
	ChannelClose = "Closed a Channel with id %v" + lineEscape

	ExchangeDeclare = "Exchange '%v' of type '%v' declared" + lineEscape

	BasicPublish = "Message to exchange '%v' with routing key '%v'" + lineEscape
	BasicBody    = "Received body:\n%v" + lineEscape
)

// handle sends answers to `message` on the provided `conn`
func (s Server) handle(message client.Message, conn net.Conn) {
	var err error

	switch msg := message.(type) {
	// connection
	case client.Hello:
		s.Log.Printf(Hello)
		_, err = conn.Write(server.ConnectionStart)
	case client.ConnectionStartOk:
		s.Log.Printf(ConnectionStartOk, msg.User, msg.Pass, msg.Mechanism)
		_, err = conn.Write(server.ConnectionTune)
	case client.ConnectionTuneOk:
		s.Log.Printf(ConnectionTuneOk)
		_, err = conn.Write(server.ConnectionOpenOk)
	case client.ConnectionOpen:
		s.Log.Printf(ConnectionOpen, msg.VirtualHost)
		// conn.Write(server.ConnectionOpenOk) // is this not according to protocol?

	// channels
	case client.ChannelOpen:
		s.Log.Printf(ChannelOpen, msg.Channel)
		_, err = conn.Write(server.ChannelOpen(msg.Channel))
	case client.ChannelClose:
		s.Log.Printf(ChannelClose, msg.Channel)
		_, err = conn.Write(server.ChannelClose(msg.Channel))

	// exchange
	case client.ExchangeDeclare:
		s.Log.Printf(ExchangeDeclare, msg.Exchange, msg.Typ)
		_, err = conn.Write(server.ExchangeDeclareOk(msg.Channel))

	// basic
	case client.BasicPublish:
		s.Log.Printf(BasicPublish, msg.Exchange, msg.RoutingKey)

	case client.Body:
		s.Log.Printf(BasicBody, msg.Payload)

	case client.Invalid:
		s.Log.Printf(msg.Err)

	// do nothing for those
	case client.Heartbeat:
	case client.Header:
	case client.Nothing:
	}

	if err != nil {
		panic(err)
	}
}

const bufferSize = 2048

func Stream(conn net.Conn) chan client.Message {
	stream := make(chan client.Message)
	go func() {
		defer close(stream)

		for {
			buf := make([]byte, bufferSize)
			size, err := conn.Read(buf)

			switch {
			case errors.Is(err, io.EOF):
				continue

			case errors.Is(err, syscall.ECONNRESET):
				fmt.Printf("Time out for %s", conn.RemoteAddr().String())
				break

			case err != nil:
				fmt.Println(err)
				break
			}

			msgDump := buf[:size]
			if bytes.Contains(msgDump, []byte{206, 178}) { // todo: ugly edge case
				stream <- amqp.Parse(msgDump)
				continue
			}

			msgs := bytes.Split(msgDump, []byte{amqp.EndMark})
			for _, msg := range msgs {
				stream <- amqp.Parse(msg)
			}
		}
	}()

	return stream
}
