// +build linux

package mqtt

import (
	"bufio"
)

type SubackPacket struct {
	FixedHeader
	MessageID []byte
	QoS       byte
}

func (subackPacket *SubackPacket) Send(conn *bufio.Writer) error {

	subackPacket.FixedHeader.RemainingLength = 3 // header + MessageID + qos

	packet := subackPacket.FixedHeader.Pack()
	packet.Write(subackPacket.MessageID)
	packet.WriteByte(subackPacket.QoS)
	_, _ = conn.Write(packet.Bytes())
	return conn.Flush()
}
