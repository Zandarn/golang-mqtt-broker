// +build linux

package mqtt

type PublishPacket struct {
	FixedHeader
	Message       string
	Topic         string
	MessageID     []byte
	Payload       []byte
	MessageLength int
}



func (publishPacket *PublishPacket) Unpack() {

	packetShift = 1
	topicLength = 0
	messageLength = 0

	//https://docs.solace.com/MQTT-311-Prtl-Conformance-Spec/MQTT%20Control%20Packets.htm#_Toc430864901

	if fixedHeader.RemainingLength > 127 {
		packetShift = 2
	} else if fixedHeader.RemainingLength > 16383 {
		packetShift = 3
	} else if fixedHeader.RemainingLength > 2097151 {
		packetShift = 4
	}
	packetShift += 2

	topicLength = int(buffer[packetShift])
	packetShift++
	publishPacket.Topic = string(buffer[packetShift : packetShift+topicLength])
	packetShift += topicLength

	if publishPacket.FixedHeader.Qos == 1 {
		publishPacket.MessageID = buffer[packetShift : packetShift+2]
	}

	messageLength = fixedHeader.RemainingLength - packetShift
	packetShift += 2

	publishPacket.Message = string(buffer[packetShift : packetShift+messageLength])
	publishPacket.MessageLength = packetShift + messageLength
	//fmt.Printf("%+v \n", publishPacket)
}
