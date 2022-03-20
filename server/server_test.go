package server

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestIntegration(t *testing.T) {
	buf := new(bytes.Buffer)
	srv := Server{
		Port: 8080,
		Log:  log.New(buf, "", log.LstdFlags),
	}
	go srv.Start()
	isPrinted(t, buf, fmt.Sprintf(Started, 8080))

	conn, err := amqp.Dial("amqp://user:pass@localhost:8080/sample-vhost")
	isNil(t, err)
	isPrinted(t, buf, Hello)
	isPrinted(t, buf, fmt.Sprintf(ConnectionStartOk, "user", "pass", "PLAIN"))
	isPrinted(t, buf, ConnectionTuneOk)

	ch, err := conn.Channel()
	isNil(t, err)
	isPrinted(t, buf, fmt.Sprintf(ChannelOpen, 1))

	err = ch.Publish("example-exchange", "my.routing.key", true, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte("Hello World"),
		})
	isNil(t, err)
	isPrinted(t, buf, fmt.Sprintf(BasicPublish, "example-exchange", "my.routing.key"))
	isPrinted(t, buf, fmt.Sprintf(BasicBody, "Hello World"))

	t.Log(buf.String())
}

func isNil(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Error(err)
	}
}

func isPrinted(t *testing.T, haystack *bytes.Buffer, needle string) {
	t.Helper()

	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Errorf("did not log '%v' within 1s\n", needle)
		default:
			if strings.Contains(haystack.String(), needle) {
				return
			}
		}
	}
}
