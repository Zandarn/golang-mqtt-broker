// +build linux

package mqtt

import (
	"bufio"
)

type PubackPacket struct {
	FixedHeader
	MessageID []byte
}

func (pubackPacket *PubackPacket) Send(conn *bufio.Writer) error {

	pubackPacket.FixedHeader.RemainingLength = 2 // header + MessageID
	packet := pubackPacket.FixedHeader.Pack()
	packet.Write(pubackPacket.MessageID)
	conn.Write(packet.Bytes())
	return conn.Flush()
}
