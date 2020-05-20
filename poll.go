// +build linux

package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"mainProject/mqtt"
	"net"
	"reflect"
	"sync"
)

type epoll struct {
	fd          int
	eventFd     int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

func createPoll() (*epoll, error) {

	fd, err := unix.EpollCreate1(0)
	panicIfErr(err)

	r0, _, errno := unix.Syscall(unix.SYS_EVENTFD2, 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	eventFd := int(r0)

	err = unix.EpollCtl(fd, unix.EPOLL_CTL_ADD, eventFd, &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(eventFd),
	})

	if err != nil {
		unix.Close(fd)
		unix.Close(eventFd)
		return nil, err
	}

	return &epoll{
		fd:          fd,
		eventFd:     eventFd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

func (poll *epoll) Add(connection net.Conn) {

	fd := getFdFromConnectionTLS(connection)

	err := unix.EpollCtl(
		poll.fd,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	panicIfErr(err)

	poll.lock.Lock()
	poll.connections[fd] = connection
	poll.lock.Unlock()
}

func (poll *epoll) Remove(connection net.Conn) {

	fd := getFdFromConnectionTLS(connection)

	err := unix.EpollCtl(poll.fd, unix.EPOLL_CTL_DEL, fd, nil)
	panicIfErr(err)
	poll.lock.Lock()
	delete(poll.connections, fd)
	poll.lock.Unlock()
	connection.Close()

}

func (poll *epoll) Wait() {

	events := make([]unix.EpollEvent, 100000)
	var numberOfEvents = 0
	var err error
	var conn net.Conn
	var broker, err1 = mqtt.CreateBroker()
	panicIfErr(err1)

	for {
		numberOfEvents, _ = unix.EpollWait(poll.fd, events, 100)

		for i := 0; i < numberOfEvents; i++ { //любое событие из пула попадает сюда
			poll.lock.RLock()
			conn = poll.connections[int(events[i].Fd)]
			poll.lock.RUnlock()
			err = broker.EventHandling(conn)

			if err == io.EOF {
				fmt.Println("eof от ", conn.LocalAddr().String())
				broker.RemoveClient(conn)
				poll.Remove(conn)

			} else if err != nil {
				fmt.Println("wtf:", err)
				broker.RemoveClient(conn)
				poll.Remove(conn)
			}
		}
	}
}

func getFdFromConnection(connection net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(connection)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func getFdFromConnectionTLS(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	tcpConn = reflect.Indirect(tcpConn.Elem())
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
