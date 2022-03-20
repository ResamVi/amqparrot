package client

type Message interface{}

// Client messages
type (
	// Hello is sent by a client to initiate
	Hello struct{}

	// ConnectionStartOk starts connection negotiation
	ConnectionStartOk struct {
		Mechanism string
		User      string
		Pass      string
	}
)

// Server messages
type (
	ConnectionTune struct{}

	ConnectionTuneOk struct {
		ChannelMax     uint16
		FrameMax       uint32
		HeartbeatDelay uint16
	}

	ConnectionOpen struct {
		VirtualHost string
	}

	ChannelOpen struct {
		Channel uint16
	}

	ChannelClose struct {
		Channel uint16
	}

	BasicPublish struct {
		Exchange   string
		RoutingKey string
	}

	Body struct {
		Payload string
	}

	ExchangeDeclare struct {
		Channel  uint16
		Exchange string
		Typ      string
	}

	Header struct{}

	Invalid struct {
		Err string
	}

	Heartbeat struct{}

	Nothing struct{}
)
