package amqp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/resamvi/amqparrot/amqp/client"
)

func Parse(msg []byte) client.Message {
	if len(msg) == 0 {
		return client.Nothing{}
	}

	if bytes.Compare(msg, Hello) == 0 {
		return client.Hello{}
	}

	buffer := bytes.NewBuffer(msg)

	var typ uint8
	if err := binary.Read(buffer, binary.BigEndian, &typ); err != nil {
		return client.Invalid{Err: "could not read type of frame"}
	}

	if typ == TypeHeartbeat || typ == TypeHeader {
		return client.Nothing{}
	}

	var (
		channel uint16
		size    uint32
	)

	if err := binary.Read(buffer, binary.BigEndian, &channel); err != nil {
		return client.Invalid{Err: "could not read channel"}
	}
	if err := binary.Read(buffer, binary.BigEndian, &size); err != nil {
		return client.Invalid{Err: err.Error()}
	}

	if typ == TypeBody {
		payload, err := io.ReadAll(buffer)
		if err != nil {
			return client.Invalid{Err: err.Error()}
		}

		return client.Body{
			Payload: string(payload),
		}
	}

	// Methods
	if typ != TypeMethod {
		return client.Invalid{Err: fmt.Sprintf("unknown type: %v", typ)}
	}

	var (
		class  uint16
		method uint16
	)

	if err := binary.Read(buffer, binary.BigEndian, &class); err != nil {
		return client.Invalid{Err: "could not read class"}
	}
	if err := binary.Read(buffer, binary.BigEndian, &method); err != nil {
		return client.Invalid{Err: "could not read method"}
	}

	// todo: bissel verschachteln?
	if class == ClassConnection && method == MethodConnectionStartOk {
		relevant := msg[140:] // todo: dirty
		buf := bytes.NewBuffer(relevant)

		var msize uint8
		if err := binary.Read(buf, binary.BigEndian, &msize); err != nil {
			return client.Invalid{Err: "could not read msize"}
		}

		mechanism := make([]byte, msize)
		if err := binary.Read(buf, binary.BigEndian, &mechanism); err != nil {
			return client.Invalid{Err: "could not read mechanism"}
		}

		var csize uint32
		if err := binary.Read(buf, binary.BigEndian, &csize); err != nil {
			return client.Invalid{Err: "could not read user+pass size"}
		}
		credentials := make([]byte, csize) // PLAIN
		if err := binary.Read(buf, binary.BigEndian, &credentials); err != nil {
			return client.Invalid{Err: "could not read user and pass"}
		}
		creds := bytes.Split(credentials, []byte{0})
		user, pass := creds[1], creds[2]

		return client.ConnectionStartOk{
			Mechanism: string(mechanism),
			User:      string(user),
			Pass:      string(pass),
		}
	}
	if class == ClassConnection && method == MethodConnectionTuneOk {
		var (
			channelMax     uint16
			frameMax       uint32
			heartbeatDelay uint16
		)
		if err := binary.Read(buffer, binary.BigEndian, &channelMax); err != nil {
			return client.Invalid{Err: "could not read channel max"}
		}
		if err := binary.Read(buffer, binary.BigEndian, &frameMax); err != nil {
			return client.Invalid{Err: "could not read frame max"}
		}
		if err := binary.Read(buffer, binary.BigEndian, &heartbeatDelay); err != nil {
			return client.Invalid{Err: "could not read heartbeat delay"}
		}

		return client.ConnectionTuneOk{
			ChannelMax:     channelMax,
			FrameMax:       frameMax,
			HeartbeatDelay: heartbeatDelay,
		}
	}

	if class == ClassConnection && method == MethodConnectionOpen {
		var vsize uint8
		if err := binary.Read(buffer, binary.BigEndian, &vsize); err != nil {
			return client.Invalid{Err: "could not read vhost size"}
		}

		vhost := make([]byte, vsize)
		if err := binary.Read(buffer, binary.BigEndian, &vhost); err != nil {
			return client.Invalid{Err: "could not read vhost"}
		}

		return client.ConnectionOpen{VirtualHost: string(vhost)}
	}

	if class == ClassChannel && method == MethodChannelOpen {
		return client.ChannelOpen{Channel: channel}
	}
	if class == ClassChannel && method == MethodChannelClose {
		return client.ChannelClose{Channel: channel}
	}

	if class == ClassExchange && method == MethodExchangeDeclare {
		var esize [3]byte // todo: this repeats below (thinking)
		if err := binary.Read(buffer, binary.BigEndian, &esize); err != nil {
			return client.Invalid{Err: "could not read exchange size"}
		}
		num := binary.BigEndian.Uint16(esize[1:])

		exchange := make([]byte, num)
		if err := binary.Read(buffer, binary.BigEndian, &exchange); err != nil {
			return client.Invalid{Err: "could not read exchange"}
		}

		var tsize uint8
		if err := binary.Read(buffer, binary.BigEndian, &tsize); err != nil {
			return client.Invalid{Err: "could not read type size"}
		}

		typ := make([]byte, tsize)
		if err := binary.Read(buffer, binary.BigEndian, &typ); err != nil {
			return client.Invalid{Err: "could not read type"}
		}

		return client.ExchangeDeclare{
			Channel:  channel,
			Exchange: string(exchange),
			Typ:      string(typ),
		}
	}

	if class == ClassBasic && method == MethodBasicPublish {
		var esize [3]byte // idk why its 3 bytes
		if err := binary.Read(buffer, binary.BigEndian, &esize); err != nil {
			return client.Invalid{Err: "could not read exchange size"}
		}
		num := binary.BigEndian.Uint16(esize[1:])

		exchange := make([]byte, num)
		if err := binary.Read(buffer, binary.BigEndian, &exchange); err != nil {
			return client.Invalid{Err: "could not read exchange"}
		}

		var rksize uint8
		if err := binary.Read(buffer, binary.BigEndian, &rksize); err != nil {
			return client.Invalid{Err: "could not read routing key size"}
		}

		routingKey := make([]byte, rksize)
		if err := binary.Read(buffer, binary.BigEndian, &routingKey); err != nil {
			return client.Invalid{Err: "could not read routing key"}
		}

		return client.BasicPublish{
			Exchange:   string(exchange),
			RoutingKey: string(routingKey),
		}
	}

	return client.Invalid{Err: "frame could not be parsed"}
}
