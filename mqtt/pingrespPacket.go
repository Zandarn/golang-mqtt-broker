// +build linux

package mqtt

import (
	"bufio"
)

type PingrespPacket struct {
	FixedHeader FixedHeader
}

func (pingrespPacket *PingrespPacket) Send(conn *bufio.Writer) error {

	pingrespPacket.FixedHeader.RemainingLength = 0
	packet := pingrespPacket.FixedHeader.Pack()
	conn.Write(packet.Bytes())
	return conn.Flush()
}
