package amqp

const (
	TypeMethod    uint8 = 1
	TypeHeader    uint8 = 2
	TypeBody      uint8 = 3
	TypeHeartbeat uint8 = 8

	GlobalChannel uint16 = 0

	ClassConnection uint16 = 10
	ClassChannel    uint16 = 20
	ClassExchange   uint16 = 40
	ClassBasic      uint16 = 60

	MethodConnectionStart   uint16 = 10
	MethodConnectionStartOk uint16 = 11
	MethodConnectionTune    uint16 = 30
	MethodConnectionTuneOk  uint16 = 31
	MethodConnectionOpen    uint16 = 40
	MethodConnectionOpenOk  uint16 = 41

	MethodChannelOpen    uint16 = 10
	MethodChannelOpenOk  uint16 = 11
	MethodChannelClose   uint16 = 40
	MethodChannelCloseOk uint16 = 41

	MethodExchangeDeclare   uint16 = 10
	MethodExchangeDeclareOk uint16 = 11

	MethodBasicPublish uint16 = 40
)

const (
	MajorVersion    uint8 = 0
	MinorVersion    uint8 = 9
	RevisionVersion uint8 = 1

	EndMark uint8 = 206
)

var (
	Hello = []byte{'A', 'M', 'Q', 'P', 0, MajorVersion, MinorVersion, RevisionVersion}
)

type (
	LongString string

	// ConnectionStart is sent by the server
	ConnectionStart struct {
		Type         uint8
		Channel      uint16
		Length       uint32
		Class        uint16
		Method       uint16
		MajorVersion uint8
		MinorVersion uint8
		MapLength    uint32
		Capabilities []string // map start
		Name         string
		Information  string
		Version      string // map end (cba to implement nested parsing)
		Auth         LongString
		Locale       LongString
	}

	// ConnectionTune is sent by the server
	ConnectionTune struct {
		Type       uint8
		Channel    uint16
		Length     uint32
		Class      uint16
		Method     uint16
		ArgLength  uint8 // cba see above
		ChannelMax uint16
		FrameMax   uint32
		Heartbeat  uint8 // 8-bit not really what spec defines *shrug*
	}

	ConnectionOpenOk struct {
		Type     uint8
		Channel  uint16
		Length   uint32
		Class    uint16
		Method   uint16
		Reserved uint8
	}

	// used for channelOpenOk, channelCloseOk, exchangeDeclareOk
	Generic struct {
		Type     uint8
		Channel  uint16
		Length   uint32
		Class    uint16
		Method   uint16
		Reserved uint32
	}
)
