// +build linux

package mqtt

import (
	"encoding/binary"
)

type UnsubscribePacket struct {
	FixedHeader
	topic     string
	MessageID []byte
}

func (unsubscribePacket *UnsubscribePacket) Unpack() {

	packetShift = 1
	topicLength = 0
	messageLength = 0

	//https://docs.solace.com/MQTT-311-Prtl-Conformance-Spec/MQTT%20Control%20Packets.htm#_Toc430864927

	if fixedHeader.RemainingLength > 127 {
		packetShift = 2
	} else if fixedHeader.RemainingLength > 16383 {
		packetShift = 3
	} else if fixedHeader.RemainingLength > 2097151 {
		packetShift = 4
	}
	packetShift++

	unsubscribePacket.MessageID = buffer[packetShift : packetShift+2]
	packetShift += 2
	topicLength = int(binary.BigEndian.Uint16(buffer[packetShift : packetShift+2]))
	packetShift += 2
	unsubscribePacket.topic = string(buffer[packetShift : packetShift+topicLength])

	//fmt.Printf("%+v \n", unsubscribePacket)
}
