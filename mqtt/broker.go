// +build linux

/*
	https://docs.solace.com/MQTT-311-Prtl-Conformance-Spec/MQTT%20Control%20Packets.htm#_Toc430864892
	http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html
*/

package mqtt

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	PacketReserved    = 0
	PacketConnect     = 1
	PacketConnack     = 2
	PacketPublish     = 3
	PacketPuback      = 4
	PacketPubrec      = 5
	PacketPubRel      = 6
	PacketPubcomp     = 7
	PacketSubscribe   = 8
	PacketSuback      = 9
	PacketUnsubscribe = 10
	PacketUnsuback    = 11
	PacketPingreq     = 12
	PacketPingresp    = 13
	PacketDisconnect  = 14
)

var numberOfMessages = 0

var brokerError error
var fixedHeader FixedHeader
var buffer = make([]byte, 8192) //dec
var packet = make([]byte, 8192)

var connect = ConnectPacket{}
var connack = ConnackPacket{}
var publish = PublishPacket{}
var puback = PubackPacket{}
var subscribe = SubscribePacket{}
var subackPacket = SubackPacket{}
var unsubscribe = UnsubscribePacket{}
var unsuback = UnsubackPacket{}
var pingresp = PingrespPacket{}

type Broker struct {
	lock         *sync.RWMutex
	clients      map[string]*Client
	topicManager TopicManager
}

var connToClient = make(map[net.Conn]string)
var broker *Broker

func CreateBroker() (*Broker, error) {

	broker = &Broker{
		topicManager: createTopicManager(),
		clients:      make(map[string]*Client),
		lock:         &sync.RWMutex{},
	}

	/*broker.topicManager.CreateTopic("/home/pc/cpu/core0")
	broker.topicManager.CreateTopic("/home/pc/cpu/core1")
	broker.topicManager.CreateTopic("/home/pc/cpu/core2")
	broker.topicManager.CreateTopic("/home/pc/cpu/core3")
	broker.topicManager.CreateTopic("/home/pc1/cpu/core0")
	broker.topicManager.CreateTopic("/home/pc1/cpu/core1")
	broker.topicManager.CreateTopic("/home/pc1/cpu/core2")
	broker.topicManager.CreateTopic("/home/pc1/cpu/core3")*/

	go printMes()
	return broker, nil
}

func printMes() {
	var lastVal = 0
	pps := 0
	for {
		pps = numberOfMessages - lastVal
		fmt.Println("кол-во сообщений:", numberOfMessages, "pps:", pps)
		lastVal = numberOfMessages
		time.Sleep(time.Second)
	}
}

func (broker *Broker) RemoveClient(conn net.Conn) {
	broker.lock.Lock()

	if broker.clients[connToClient[conn]] != nil {
		for _, v := range broker.clients[connToClient[conn]].subscriptions {
			v.online = false
		}
		broker.clients[connToClient[conn]].connBuff.Reset(conn)
		delete(connToClient, conn)
	}

	go broker.topicManager.ClearUnusedTopics()
	broker.lock.Unlock()

}

func (broker *Broker) AddClient(conn net.Conn) {

	connToClient[conn] = connect.ClientIdentifier

	if broker.clients[connect.ClientIdentifier] != nil {
		broker.clients[connect.ClientIdentifier].connection = conn
		broker.clients[connect.ClientIdentifier].connBuff = bufio.NewWriter(conn)
	} else {
		broker.clients[connect.ClientIdentifier] = &Client{
			lock:          &sync.RWMutex{},
			id:            connect.ClientIdentifier,
			connBuff:      bufio.NewWriter(conn),
			connection:    conn,
			username:      connect.Username,
			password:      connect.Password,
			subscriptions: make(map[string]*Subscription),
		}
	}
}

func (broker *Broker) EventHandling(conn net.Conn) error {

	if conn == nil {
		return nil
	}
	_, brokerError = conn.Read(buffer)
	if brokerError != nil {
		return brokerError
	}

	fixedHeader.Unpack()

	switch fixedHeader.MessageType {

	case PacketReserved:
		{
			fmt.Println("Reserved. Wtf???")
		}
	case PacketConnect:
		{
			brokerError = broker.ReceiveConnect(conn)
			if brokerError != nil {
				return brokerError
			}
			break
		}
	case PacketConnack:
		{
			//fmt.Println("Connack")
		}
	case PacketPublish:
		{
			brokerError = broker.ReceivePublish(conn)
			if brokerError != nil {
				return brokerError
			}
		}
	case PacketPuback:
		{
			brokerError = broker.ReceivePuback(conn)
			if brokerError != nil {
				return brokerError
			}
		}
	case PacketSubscribe:
		{
			brokerError = broker.ReceiveSubscribe(conn)
			if brokerError != nil {
				return brokerError
			}
		}
	case PacketUnsubscribe:
		{
			brokerError = broker.ReceiveUnsubscribe(conn)
			if brokerError != nil {
				return brokerError
			}
		}
	case PacketPingreq:
		{
			brokerError = broker.ReceivePingreq(conn)
			if brokerError != nil {
				return brokerError
			}
		}
	case PacketDisconnect:
		{
			return io.EOF
		}
	default:
		{
			fmt.Println("default. Unknown message type:", fixedHeader.MessageType)
		}
	}

	return brokerError
}

func (broker *Broker) ReceiveConnect(conn net.Conn) error {
	connect.Unpack()

	if connect.UsernameFlag && connect.PasswordFlag {
		fmt.Println("проверка логина и пароля")
	}

	brokerError = broker.SendConnack(conn)
	if brokerError != nil {
		return brokerError
	}

	return brokerError
}

func (broker *Broker) SendConnack(conn net.Conn) error {
	connack.FixedHeader = fixedHeader
	connack.FixedHeader.MessageType = PacketConnack
	connack.SessionPresent = false
	connack.ReturnCode = 0

	broker.AddClient(conn)
	brokerError = connack.Send(broker.clients[connToClient[conn]].connBuff)
	if brokerError != nil {
		return brokerError
	}

	return nil
}

func (broker *Broker) ReceivePublish(conn net.Conn) error {
	publish.FixedHeader = fixedHeader
	publish.Unpack()

	copy(packet, buffer[:publish.MessageLength])

	brokerError = broker.topicManager.SendMessageToSubscribers(publish.Topic, publish.FixedHeader.Retain, packet[:publish.MessageLength])
	if brokerError != nil {
		return brokerError
	}

	if publish.Qos == 1 {
		brokerError = broker.SendPuback(conn)
		if brokerError != nil {
			return brokerError
		}
	} else if publish.Qos == 2 {
		brokerError = broker.SendPubrec(conn)
		if brokerError != nil {
			return brokerError
		}
	}

	return nil
}

func (broker *Broker) ReceivePuback(conn net.Conn) error {
	fmt.Println("TODO получение Puback")
	return nil
}

func (broker *Broker) SendPuback(conn net.Conn) error {

	puback.FixedHeader = fixedHeader
	puback.MessageID = publish.MessageID
	puback.FixedHeader.MessageType = PacketPuback

	brokerError = puback.Send(broker.clients[connToClient[conn]].connBuff)
	return brokerError
}

func (broker *Broker) SendPubrec(conn net.Conn) error {
	fmt.Println("TODO отправка PUBREC") //TODO
	return nil
}

func (broker *Broker) ReceivePubrel(conn net.Conn) error {

	fmt.Println("TODO отправка PUBCOMP") //TODO
	return nil
}

func (broker *Broker) ReceiveSubscribe(conn net.Conn) error {

	subscribe.FixedHeader = fixedHeader
	subscribe.Unpack()

	brokerError = broker.topicManager.Subscribe(broker.clients[connToClient[conn]], subscribe.topic)
	if brokerError != nil {
		return brokerError
	}

	for _, v := range broker.clients[connToClient[conn]].subscriptions {
		go broker.topicManager.SendMessagesToSubscriber(broker.clients[connToClient[conn]], v)
	}

	return broker.SendSuback(conn)
}

func (broker *Broker) SendSuback(conn net.Conn) error {

	subackPacket.FixedHeader = fixedHeader
	subackPacket.FixedHeader.MessageType = PacketSuback
	subackPacket.MessageID = subscribe.MessageID
	subackPacket.QoS = subscribe.QoS
	return subackPacket.Send(broker.clients[connToClient[conn]].connBuff)
}

func (broker *Broker) ReceiveUnsubscribe(conn net.Conn) error {

	unsubscribe.Unpack()
	broker.topicManager.Unsubscribe(broker.clients[connToClient[conn]], unsubscribe.topic)
	brokerError = broker.SendUnsuback(conn)
	return brokerError
}

func (broker *Broker) SendUnsuback(conn net.Conn) error {

	unsuback.FixedHeader = fixedHeader
	unsuback.FixedHeader.MessageType = PacketUnsuback
	unsuback.MessageID = subscribe.MessageID
	return unsuback.Send(broker.clients[connToClient[conn]].connBuff)
}

func (broker *Broker) ReceivePingreq(conn net.Conn) error {

	brokerError = broker.SendPingresp(conn)
	return brokerError
}

func (broker *Broker) SendPingresp(conn net.Conn) error {

	pingresp.FixedHeader = fixedHeader
	pingresp.FixedHeader.MessageType = PacketPingresp
	return pingresp.Send(broker.clients[connToClient[conn]].connBuff)
}
