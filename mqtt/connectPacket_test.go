// +build linux

package mqtt

import (
	"testing"
)

func TestConnectPacket(t *testing.T) {

	connect = ConnectPacket{}
	buffer = []byte{16, 52, 0, 4, 77, 81, 84, 84, 4, 204, 0, 0, 0, 0, 0, 4, 116, 101, 115, 116, 0, 12, 84, 101, 115, 116, 32, 80, 97, 121, 108, 111, 97, 100, 0, 8, 116, 101, 115, 116, 117, 115, 101, 114, 0, 8, 116, 101, 115, 116, 112, 97, 115, 115}
	connect.Unpack()

	if connect.ProtocolName != "MQTT" {
		t.Errorf("Connect Packet ProtocolName is %s, should be %s", connect.ProtocolName, "MQTT")
	}
	if connect.ProtocolVersion != 4 {
		t.Errorf("Connect Packet ProtocolVersion is %d, should be %d", connect.ProtocolVersion, 4)
	}
	if connect.UsernameFlag != true {
		t.Errorf("Connect Packet UsernameFlag is %t, should be %t", connect.UsernameFlag, true)
	}
	if connect.Username != "testuser" {
		t.Errorf("Connect Packet Username is %s, should be %s", connect.Username, "testuser")
	}
	if connect.PasswordFlag != true {
		t.Errorf("Connect Packet PasswordFlag is %t, should be %t", connect.PasswordFlag, true)
	}
	if string(connect.Password) != "testpass" {
		t.Errorf("Connect Packet Password is %s, should be %s", string(connect.Password), "testpass")
	}
	if connect.WillFlag != true {
		t.Errorf("Connect Packet WillFlag is %t, should be %t", connect.WillFlag, true)
	}
	if string(connect.WillTopic) != "test" {
		t.Errorf("Connect Packet WillTopic is %s, should be %s", string(connect.WillTopic), "test")
	}
	if connect.WillQos != 1 {
		t.Errorf("Connect Packet WillQos is %d, should be %d", connect.WillQos, 1)
	}
	if connect.WillRetain != false {
		t.Errorf("Connect Packet WillRetain is %t, should be %t", connect.WillRetain, false)
	}
	if string(connect.WillMessage) != "Test Payload" {
		t.Errorf("Connect Packet WillMessage is %s, should be %s", string(connect.WillMessage), "Test Payload")
	}
}
