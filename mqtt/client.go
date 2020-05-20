// +build linux

package mqtt

import (
	"bufio"
	"net"
	"sync"
)

type Client struct {
	lock          *sync.RWMutex
	id            string
	connection    net.Conn
	connBuff      *bufio.Writer
	username      string                   // todo нужно ли вообще
	password      []byte                   // todo нужно ли вообще
	subscriptions map[string]*Subscription // ключ - имя топика
}

type Subscription struct {
	node    *Node
	lock    *sync.RWMutex
	QoS     byte
	online  bool
	storage map[int64][]byte
}
