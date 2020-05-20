// +build linux

package mqtt

import (
	"encoding/binary"
)

type ConnectPacket struct {
	FixedHeader
	ProtocolName    string
	ProtocolVersion byte
	CleanSession    bool
	WillFlag        bool
	WillQos         byte
	WillRetain      bool
	UsernameFlag    bool
	PasswordFlag    bool
	ReservedBit     byte
	Keepalive       uint16

	ClientIdentifier string
	WillTopic        []byte
	WillMessage      []byte
	Username         string
	Password         []byte
}

var packetShift = 1
var topicLength = 0
var messageLength = 0

func (connectPacket *ConnectPacket) Unpack() {

	var shift uint16 = 1
	if fixedHeader.RemainingLength > 127 {
		shift = 2
	} else if fixedHeader.RemainingLength > 16383 {
		shift = 3
	} else if fixedHeader.RemainingLength > 2097151 {
		shift = 4
	}
	shift += 2

	protocolLength := uint16(buffer[shift])
	shift++
	connectPacket.ProtocolName = string(buffer[shift : shift+protocolLength])
	shift += protocolLength
	connectPacket.ProtocolVersion = buffer[shift]
	shift++
	connectFlags := buffer[shift]
	connectPacket.ReservedBit = 1 & connectFlags
	connectPacket.CleanSession = 1&(connectFlags>>1) > 0
	connectPacket.WillFlag = 1&(connectFlags>>2) > 0
	connectPacket.WillQos = 3 & (connectFlags >> 3)
	connectPacket.WillRetain = 1&(connectFlags>>5) > 0
	connectPacket.PasswordFlag = 1&(connectFlags>>6) > 0
	connectPacket.UsernameFlag = 1&(connectFlags>>7) > 0
	shift++
	connectPacket.Keepalive = binary.BigEndian.Uint16(buffer[shift : shift+2])
	shift += 2
	clientIDLength := binary.BigEndian.Uint16(buffer[shift : shift+2])
	shift += 2
	connectPacket.ClientIdentifier = string(buffer[shift : shift+clientIDLength])
	shift += clientIDLength

	if connectPacket.WillFlag {
		willTopicLength := binary.BigEndian.Uint16(buffer[shift : shift+2])
		shift += 2
		connectPacket.WillTopic = buffer[shift : shift+willTopicLength] //TODO сбрасывать коннект, если топик не правильный (путь) ?
		shift += willTopicLength
		willMessageLength := binary.BigEndian.Uint16(buffer[shift : shift+2])
		shift += 2
		connectPacket.WillMessage = buffer[shift : shift+willMessageLength]
		shift += willMessageLength
	}

	if connectPacket.UsernameFlag {
		userNameLength := binary.BigEndian.Uint16(buffer[shift : shift+2])
		shift += 2
		connectPacket.Username = string(buffer[shift : shift+userNameLength])
		shift += userNameLength
	}

	if connectPacket.PasswordFlag {
		passwordLength := binary.BigEndian.Uint16(buffer[shift : shift+2])
		shift += 2
		connectPacket.Password = buffer[shift : shift+passwordLength]
		shift += passwordLength
	}
	//fmt.Printf("%+v \n", connectPacket)
}
