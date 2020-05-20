// +build linux

package mqtt

import (
	"bufio"
	"reflect"
	"sync"
	"testing"
)

func TestCreateTopicManager(t *testing.T) {

	topicManager := createTopicManager()

	if reflect.TypeOf(topicManager.topics).Kind().String() != "map" {
		t.Error("В topicManager нет мапы topics")
	}

	if reflect.TypeOf(topicManager.roots).Kind().String() != "map" {
		t.Error("В topicManager нет мапы roots")
	}
}

func TestCreateTopic(t *testing.T) {
	topicManager := createTopicManager()

	_ = topicManager.CreateTopic("/home/pc/cpu/core0")
	_ = topicManager.CreateTopic("/home/pc1/cpu/core0")
	_ = topicManager.CreateTopic("/home/pc/cpu/core1")
	_ = topicManager.CreateTopic("/home/pc1/cpu/core1")

	if topicManager.roots["home"] == nil {
		t.Errorf("Нет корня home\n")
	}
	if topicManager.roots["home"].nodes["pc"] == nil {
		t.Errorf("Нет ноды /home/pc\n")
	}
	if topicManager.roots["home"].nodes["pc1"] == nil {
		t.Errorf("Нет ноды /home/pc1\n")
	}
	if topicManager.roots["home"].nodes["pc"].nodes["cpu"] == nil {
		t.Errorf("Нет ноды /home/pc/cpu\n")
	}
	if topicManager.roots["home"].nodes["pc1"].nodes["cpu"] == nil {
		t.Errorf("Нет ноды /home/pc1/cpu\n")
	}
	if topicManager.roots["home"].nodes["pc"].nodes["cpu"].nodes["core0"] == nil {
		t.Errorf("Нет ноды /home/pc/cpu/core0\n")
	}
	if topicManager.roots["home"].nodes["pc"].nodes["cpu"].nodes["core1"] == nil {
		t.Errorf("Нет ноды /home/pc/cpu/core1\n")
	}
	if topicManager.roots["home"].nodes["pc1"].nodes["cpu"].nodes["core0"] == nil {
		t.Errorf("Нет ноды /home/pc/cpu/core0\n")
	}
	if topicManager.roots["home"].nodes["pc1"].nodes["cpu"].nodes["core1"] == nil {
		t.Errorf("Нет ноды /home/pc1/cpu/core1\n")
	}

	topicManager.CreateSubTopic(topicManager.roots["home"].nodes["pc"].nodes["cpu"], "/home/pc/cpu/", "core4")
	if topicManager.roots["home"].nodes["pc"].nodes["cpu"].nodes["core4"] == nil {
		t.Errorf("Нет ноды /home/pc/cpu/core4\n")
	}
}

func TestCreateTopicForMaskSubscribers(t *testing.T) {
	topicManager := createTopicManager()
	client := &Client{
		lock:          &sync.RWMutex{},
		id:            connect.ClientIdentifier,
		connBuff:      bufio.NewWriter(nil),
		connection:    nil,
		username:      "test",
		password:      make([]byte, 1),
		subscriptions: make(map[string]*Subscription),
	}
	maskSubscribers["/home/+/cpu/+"] = append(maskSubscribers["/home/+/cpu/+"], client)

	_ = topicManager.CreateTopicForMaskSubscribers("/home/pc/cpu/core0")

	if topicManager.roots["home"] == nil {
		t.Errorf("Нет корня home\n")
	}
	if topicManager.roots["home"].nodes["pc"] == nil {
		t.Errorf("Нет ноды /home/pc\n")
	}
	if topicManager.roots["home"].nodes["pc"].nodes["cpu"] == nil {
		t.Errorf("Нет ноды /home/pc/cpu\n")
	}
	if topicManager.roots["home"].nodes["pc"].nodes["cpu"].nodes["core0"] == nil {
		t.Errorf("Нет ноды /home/pc/cpu/core0\n")
	}
}

func TestSubscribeOnExistingTopics(t *testing.T) {
	topicManager := createTopicManager()

	client := &Client{
		lock:          &sync.RWMutex{},
		id:            connect.ClientIdentifier,
		connBuff:      bufio.NewWriter(nil),
		connection:    nil,
		username:      "test",
		password:      make([]byte, 1),
		subscriptions: make(map[string]*Subscription),
	}

	_ = topicManager.CreateTopic("/home/pc/cpu/core0")
	_ = topicManager.CreateTopic("/home/pc1/cpu/core0")
	maskSubscribers["/home/+/cpu/+"] = append(maskSubscribers["/home/+/cpu/+"], client)
	topicManager.SubscribeOnExistingTopics(client, "/home/+/cpu/+")

	if client.subscriptions["/home/pc/cpu/core0"] == nil {
		t.Error("Нет подписки на /home/pc/cpu/core0")
	}

	if client.subscriptions["/home/pc1/cpu/core0"] == nil {
		t.Error("Нет подписки на /home/pc1/cpu/core0")
	}
}

func TestUnsubscribe(t *testing.T) {
	topicManager := createTopicManager()

	client := &Client{
		lock:          &sync.RWMutex{},
		id:            connect.ClientIdentifier,
		connBuff:      bufio.NewWriter(nil),
		connection:    nil,
		username:      "test",
		password:      make([]byte, 1),
		subscriptions: make(map[string]*Subscription),
	}

	_ = topicManager.CreateTopic("/home/pc/cpu/core0")
	_ = topicManager.CreateTopic("/home/pc1/cpu/core0")
	maskSubscribers["/home/+/cpu/+"] = append(maskSubscribers["/home/+/cpu/+"], client)
	topicManager.SubscribeOnExistingTopics(client, "/home/+/cpu/+")

	topicManager.Unsubscribe(client, "/home/+/cpu/+")

	if len(client.subscriptions) != 0 {
		t.Error("Не пустые подписки")
	}

}
