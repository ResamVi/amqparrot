package server

import (
	"github.com/resamvi/amqparrot/amqp"
	"reflect"
	"strings"
)

var (
	// ConnectionStart is the answer to "hello" sent to client
	ConnectionStart []byte
	// ConnectionTune is the answer to "ConnectionStartOk" sent to client
	ConnectionTune []byte
	// ConnectionOpenOk is the answer to "ConnectionTuneOk" sent to client
	ConnectionOpenOk []byte
)

func init() {
	ConnectionStart = MarshalBinary(amqp.ConnectionStart{
		Type:         amqp.TypeMethod,
		Channel:      amqp.GlobalChannel,
		Length:       0, // rewritten later
		Class:        amqp.ClassConnection,
		Method:       amqp.MethodConnectionStart,
		MajorVersion: 0,
		MinorVersion: 9,
		MapLength:    327, // hardcoded size: I cba to implement nested parsing to calculate size of payload
		Capabilities: []string{
			"publisher_confirms",
			"exchange_exchange_bindings",
			"basic.nack",
			"consumer_cancel_notify",
			"connection.blocked",
			"consumer_priorities",
			"authentication_failure_close",
			"per_consumer_qos",
			"direct_reply_to",
		},
		Name:        "amqparrot",
		Information: "MIT License - Copyright (c) 2022 Julien Midedji",
		Version:     "1.0.0 (go1.18)",
		Auth:        "AMQPLAIN PLAIN",
		Locale:      "en_US",
	})

	ConnectionTune = MarshalBinary(amqp.ConnectionTune{
		Type:       amqp.TypeMethod,
		Channel:    amqp.GlobalChannel,
		Length:     0, // rewritten later
		Class:      amqp.ClassConnection,
		Method:     amqp.MethodConnectionTune,
		ArgLength:  7, // hardcoded size
		ChannelMax: 1<<16 - 256,
		FrameMax:   1 << 25,
		Heartbeat:  60,
	})

	ConnectionOpenOk = MarshalBinary(amqp.ConnectionOpenOk{
		Type:     amqp.TypeMethod,
		Channel:  amqp.GlobalChannel,
		Length:   0, // rewritten later
		Class:    amqp.ClassConnection,
		Method:   amqp.MethodConnectionOpenOk,
		Reserved: 0,
	})
}

func ChannelOpen(channel uint16) []byte {
	return MarshalBinary(amqp.Generic{
		Type:    amqp.TypeMethod,
		Channel: channel,
		Length:  0, // rewritten later
		Class:   amqp.ClassChannel,
		Method:  amqp.MethodChannelOpenOk,
	})
}

func ChannelClose(channel uint16) []byte {
	return MarshalBinary(amqp.Generic{
		Type:     amqp.TypeMethod,
		Channel:  channel,
		Length:   0, // rewritten later
		Class:    amqp.ClassChannel,
		Method:   amqp.MethodChannelCloseOk,
		Reserved: 0,
	})
}

func ExchangeDeclareOk(channel uint16) []byte {
	return MarshalBinary(amqp.Generic{
		Type:     amqp.TypeMethod,
		Channel:  channel,
		Length:   0, // rewritten later
		Class:    amqp.ClassExchange,
		Method:   amqp.MethodExchangeDeclareOk,
		Reserved: 0,
	})
}

// MarshalBinary converts the go structs that model amqp frames to a byte slice
// Uses all kinds of lazy tricks
func MarshalBinary[T any](strct T) []byte {
	var data []byte

	s := reflect.ValueOf(strct)
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)

		switch value := field.Interface().(type) {
		case uint8:
			data = append(data, value)
		case uint16:
			data = append(data, byte(value>>8), byte(value))
		case uint32:
			data = append(data, byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
		case amqp.LongString:
			valueLength := len(value)
			data = append(data, byte(valueLength>>24), byte(valueLength>>16), byte(valueLength>>8), byte(valueLength))
			data = append(data, []byte(value)...)

		// not really a correct implementation of the amqp protocol
		case []string: // for map of booleans only
			name := strings.ToLower(s.Type().Field(i).Name)
			data = append(data, byte(len(name)))
			data = append(data, []byte(name)...)
			data = append(data, 'F')

			m := make([]byte, 0)
			for _, name := range value {
				m = append(m, byte(len(name)))
				m = append(m, []byte(name)...)
				m = append(m, 't')
				m = append(m, 1)
			}
			mapLength := len(m)
			data = append(data, byte(mapLength>>24), byte(mapLength>>16), byte(mapLength>>8), byte(mapLength))
			data = append(data, m...)
		case string: // for fields in map only
			// name of field
			name := strings.ToLower(s.Type().Field(i).Name)
			data = append(data, byte(len(name)))
			data = append(data, []byte(name)...)
			data = append(data, 'S')

			// value of field
			valueLength := len(value)
			data = append(data, byte(valueLength>>24), byte(valueLength>>16), byte(valueLength>>8), byte(valueLength))
			data = append(data, []byte(value)...)
		}
	}

	totalLength := len(data[7:]) // location of payload length
	totalLengthBytes := []byte{byte(totalLength >> 24), byte(totalLength >> 16), byte(totalLength >> 8), byte(totalLength)}
	copy(data[3:7], totalLengthBytes)

	data = append(data, amqp.EndMark)

	return data
}
