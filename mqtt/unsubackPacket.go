// +build linux

package mqtt

import (
	"bufio"
)

type UnsubackPacket struct {
	FixedHeader
	MessageID []byte
}

func (unsubackPacket *UnsubackPacket) Send(conn *bufio.Writer) error {

	unsubackPacket.FixedHeader.RemainingLength = 2 // header + MessageID

	packet := unsubackPacket.FixedHeader.Pack()
	packet.Write(unsubackPacket.MessageID)
	conn.Write(packet.Bytes())
	return conn.Flush()
}
