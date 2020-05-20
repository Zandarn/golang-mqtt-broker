// +build linux

package mqtt

import (
	"bufio"
)

type ConnackPacket struct {
	FixedHeader
	SessionPresent bool
	ReturnCode     byte
}

func (ca *ConnackPacket) Send(conn *bufio.Writer) error {

	ca.FixedHeader.RemainingLength = 2
	packet := ca.FixedHeader.Pack()
	packet.WriteByte(boolToByte(ca.SessionPresent))
	packet.WriteByte(ca.ReturnCode)
	_, _ = conn.Write(packet.Bytes())
	return conn.Flush()
}
