// +build linux

package mqtt

import (
	"bytes"
)

/**
https://ipc2u.ru/articles/prostye-resheniya/chto-takoe-mqtt/
QoS 0
	На этом уровне издатель один раз отправляет сообщение брокеру и не ждет подтверждения от него, то есть отправил и забыл.
QoS 1
	Этот уровень гарантирует, что сообщение точно будет доставлено брокеру, но есть вероятность дублирования сообщений от издателя.
	После получения дубликата сообщения, брокер снова рассылает это сообщение подписчикам, а издателю снова отправляет подтверждение о получении сообщения.
	Если издатель не получил PUBACK сообщения от брокера, он повторно отправляет этот пакет, при этом в DUP устанавливается «1».
QoS 2
	Издатель отправляет сообщение брокеру. В этом сообщении указывается уникальный Packet ID, QoS=2 и DUP=0.
	Издатель хранит сообщение неподтвержденным пока не получит от брокера ответ PUBREC. Брокер отвечает сообщением PUBREC в котором содержится тот же Packet ID.
	После его получения издатель отправляет PUBREL с тем же Packet ID. До того, как брокер получит PUBREL он должен хранить копию сообщения у себя.
	После получения PUBREL он удаляет копию сообщения и отправляет издателю сообщение PUBCOMP о том, что транзакция завершена.
*/

var typeAndFlags byte
var decodeLengthValue = 0
var decodeLengthCounter = 0

type FixedHeader struct {
	MessageType     byte
	Dup             bool
	Qos             byte
	Retain          bool
	RemainingLength int
}

func (fh *FixedHeader) Unpack() {
	typeAndFlags = buffer[0]

	fh.MessageType = typeAndFlags >> 4
	fh.Dup = (typeAndFlags>>3)&0x01 > 0
	fh.Qos = (typeAndFlags >> 1) & 0x03
	fh.Retain = typeAndFlags&0x01 > 0
	fh.RemainingLength = decodeLength()
}

func encodeLength(length int) []byte {
	var encLength []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encLength = append(encLength, digit)
		if length == 0 {
			break
		}
	}
	return encLength
}

func decodeLength() int {

	decodeLengthValue = 0
	decodeLengthCounter = 0
	for _, data := range buffer[1:] {

		encodedByte := int(data)
		decodeLengthValue |= (encodedByte & 0x7F) << (7 * decodeLengthCounter)

		if decodeLengthCounter >= 4 {
			panic("Malformed Remaining Length")
		}

		if encodedByte&0x80 == 0 {
			break
		} else {
			decodeLengthCounter++
		}
	}

	return decodeLengthValue
}

func (fh *FixedHeader) Pack() bytes.Buffer {
	var header bytes.Buffer
	header.WriteByte(fh.MessageType<<4 | boolToByte(fh.Dup)<<4 | fh.Qos<<2 | boolToByte(fh.Retain))
	header.Write(encodeLength(fh.RemainingLength))
	return header
}

func boolToByte(b bool) byte {
	switch b {
	case true:
		return 1
	default:
		return 0
	}
}
