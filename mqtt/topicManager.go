// +build linux

package mqtt

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type TopicManager struct {
	topics map[string]*Node
	roots  map[string]*Node
}

type Node struct {
	name        string
	subscribers []string
	qos         byte
	nodes       map[string]*Node // ключ - имя топика
}

var topicManagerError error
var sendMessageClientID = "757575757575757575757575757575757575"

var maskSubscribers = make(map[string][]*Client) // ключ - имя топика с маской

func createTopicManager() TopicManager {

	topicManager := TopicManager{topics: make(map[string]*Node), roots: make(map[string]*Node)}
	return topicManager
}

func (topicManager *TopicManager) SendMessagesToSubscriber(client *Client, topic *Subscription) {

	buff := bufio.NewWriter(client.connection)
	for timestamp, packet := range topic.storage {

		_, _ = buff.Write(packet)
		err := buff.Flush()
		if err == nil {
			delete(topic.storage, timestamp)
			numberOfMessages++
		} else {
			fmt.Println("ошибка отправки для ", client.id, err)
		}
	}
	buff = nil
}

//В publish пакетах нельзя отправлять сообщения с маской в топике
func (topicManager *TopicManager) SendMessageToSubscribers(topicName string, retain bool, packet []byte) error {

	if topicManager.topics[topicName] == nil {
		fmt.Println("CreateTopicForMaskSubscribers ",topicName)
		topicManagerError = topicManager.CreateTopicForMaskSubscribers(topicName)

		if topicManagerError != nil {
			return topicManagerError
		}
	}

	if topicManager.topics[topicName] == nil {
		return errors.New("нет такого топика")
	}

	for _, sendMessageClientID = range topicManager.topics[topicName].subscribers {

		if broker.clients[sendMessageClientID].subscriptions[topicName].online {

			broker.clients[sendMessageClientID].lock.Lock()

			_, topicManagerError = broker.clients[sendMessageClientID].connBuff.Write(packet)
			if topicManagerError != nil {
				return topicManagerError
			}
			topicManagerError = broker.clients[sendMessageClientID].connBuff.Flush()

			broker.clients[sendMessageClientID].lock.Unlock()

			if topicManagerError != nil && retain {
				fmt.Println(sendMessageClientID, "Добавление сообщения в хранилище. Топик:", topicName)
				broker.clients[sendMessageClientID].subscriptions[topicName].storage[time.Now().UnixNano()/1000000] = packet
				//fmt.Println("SEND ERROR:", err)
			} else {
				numberOfMessages++
			}
		} else {
			if retain {
				fmt.Println(sendMessageClientID, " отключён. Добавление сообщения в хранилище. Топик:", topicName)
				broker.clients[sendMessageClientID].subscriptions[topicName].storage[time.Now().UnixNano()/1000000] = packet
			}
		}
	}

	packet = nil
	return nil
}

func (topicManager *TopicManager) CreateTopic(name string) error {

	splitedPath := strings.Split(name, "/")
	if splitedPath[0] == "" {
		splitedPath = splitedPath[1:]
	}

	var root = &Node{}
	if topicManager.roots[splitedPath[0]] == nil {
		root = &Node{name: splitedPath[0], nodes: make(map[string]*Node)}
		topicManager.roots[root.name] = root
	} else {
		root = topicManager.roots[splitedPath[0]]
	}

	lastNode := root

	for key, name := range splitedPath[1:] {

		if name == "+" || name == "#" {
			return errors.New("не могу создать топик(" + name + ") с маской")
		}

		var newNode = &Node{}

		if lastNode.nodes[name] == nil {
			newNode = &Node{name: name, nodes: make(map[string]*Node)}
			topicManager.topics["/"+strings.Join(splitedPath[0:key+1], "/")+"/"+name] = newNode //ключ - полный путь
			lastNode.nodes[name] = newNode
		} else {
			newNode = lastNode.nodes[name]
		}

		lastNode = newNode
	}
	/*	fmt.Println(topicManager.roots["home"].nodes["pc"].nodes["cpu"])*/
	return nil
}

func (topicManager *TopicManager) CreateSubTopic(parent *Node, fullpath string, name string) {

	parent.nodes[name] = &Node{name: name, nodes: make(map[string]*Node)}
	topicManager.topics[fullpath+"/"+name] = parent.nodes[name] //ключ - полный путь
}

func (topicManager *TopicManager) CreateTopicForMaskSubscribers(topicName string) error {

	topicName = strings.TrimRight(topicName, "/")

	regexMask := ""
	for mask := range maskSubscribers {

		if strings.Contains(mask, "/+") {

			regexMask = strings.Replace(mask, "/+", "\\/.*", -1)
			validTopic := regexp.MustCompile("^" + regexMask)

			if validTopic.MatchString(topicName) {

				err := topicManager.CreateTopic(topicName)
				if err != nil {
					validTopic = nil
					return err
				}

				validTopic = nil

			} else {
				validTopic = nil
				return errors.New("нет подписчиков для этого топика")
			}

		} else if strings.Contains(mask, "/#") {

			splitedMask := strings.Split(mask, "/")
			splitedTopic := strings.Split(topicName, "/")

		Loop:
			for k, part := range splitedTopic {

				if part != splitedMask[k] && splitedMask[k] == "#" {
					err := topicManager.CreateTopic(topicName)
					if err != nil {
						return err
					}
					break Loop
				}
			}

		}
	}

	err := topicManager.SubscribeExistingMaskClients(topicName)
	if err != nil {
		return err
	}

	return nil
}

func (topicManager *TopicManager) SubscribeOnExistingTopics(client *Client, mask string) {

	splitedPath := strings.Split(mask, "/")

	if splitedPath[0] == "" {
		splitedPath = splitedPath[1:]
	}

	if topicManager.roots[splitedPath[0]] == nil {
		return
	}

	node := topicManager.roots[splitedPath[0]]
	fullpath := "/" + topicManager.roots[splitedPath[0]].name // "/" + root и получается вроде /home
	join := ""

Loop:
	for key, name := range splitedPath[1:] {

		switch name {
		case "+":
			{
				for _, topic := range node.nodes {
					join = strings.Join(splitedPath[key+2:], "/")
					topicManager.SubscribeOnExistingTopics(client, strings.TrimRight(fullpath+"/"+topic.name+"/"+join, "/"))
				}
				break Loop
			}
		case "#":
			{
				for _, topic := range node.nodes {
					join = strings.Join(splitedPath[key+2:], "/")
					topicManager.RecursiveSubscribeOnExistingTopics(client, topic, strings.TrimRight(fullpath+"/"+topic.name+"/"+join, "/"))
				}
				break Loop
			}
		default:
			if node.nodes[name] == nil {
				continue Loop
			}

			fullpath += "/" + name
			node = node.nodes[name]

			if key == len(splitedPath)-2 {
				topicManager.AddSub(client, node, fullpath)
			}
		}
	}

}

func (topicManager *TopicManager) RecursiveSubscribeOnExistingTopics(client *Client, node *Node, fullpath string) {

	if len(node.nodes) > 0 {
		for _, topic := range node.nodes {
			topicManager.RecursiveSubscribeOnExistingTopics(client, topic, fullpath+"/"+topic.name)
		}
	} else {
		topicManager.AddSub(client, node, fullpath)
	}
}

func (topicManager *TopicManager) SubscribeExistingMaskClients(topicName string) error {

	for _, clients := range maskSubscribers {
		for _, client := range clients {
			err := topicManager.Subscribe(client, topicName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (topicManager *TopicManager) Subscribe(client *Client, topicName string) error {

	topicName = strings.TrimRight(topicName, "/")

	if strings.Contains(topicName, "/+") || strings.Contains(topicName, "/#") {

		if strings.Contains(topicName, "#/") {
			return errors.New("не правильная маска " + topicName)
		}

		isIn := func() bool {
			for _, v := range maskSubscribers[topicName] {
				if v.id == client.id {
					return true
				}
			}
			return false
		}()

		if !isIn {
			maskSubscribers[topicName] = append(maskSubscribers[topicName], client)
		}

		topicManager.SubscribeOnExistingTopics(client, topicName)
		return nil
	}

	splitedPath := strings.Split(topicName, "/")

	if splitedPath[0] == "" {
		splitedPath = splitedPath[1:]
	}

	if topicManager.roots[splitedPath[0]] == nil {
		err := topicManager.CreateTopic(topicName)
		if err != nil {
			return err
		}
	}

	node := topicManager.roots[splitedPath[0]]
	fullpath := "/" + topicManager.roots[splitedPath[0]].name // "/" + root и получается вроде /home

	for key, name := range splitedPath[1:] {

		fullpath += "/" + name

		if topicManager.topics[fullpath] == nil || node.nodes[name] == nil {
			topicManager.CreateSubTopic(node, fullpath, name)
		}

		node = node.nodes[name]

		if key == len(splitedPath)-2 {
			topicManager.AddSub(client, node, fullpath)
		}
	}

	return nil
}

func (topicManager *TopicManager) AddSub(client *Client, topic *Node, fullName string) {

	if client.subscriptions[fullName] == nil {
		client.subscriptions[fullName] = &Subscription{
			online:  true,
			node:    topic,
			QoS:     subscribe.QoS,
			storage: make(map[int64][]byte),
		}

		//ссылку не желательно совать, т.к. append может переместить всё в другой участок памяти и слайс от subscribers сломается
		topic.subscribers = append(topic.subscribers, client.id)
	} else {
		client.subscriptions[fullName].online = true
	}
}

func (topicManager *TopicManager) Unsubscribe(client *Client, topicName string) {

	topicName = strings.TrimRight(topicName, "/")

	if strings.Contains(topicName, "/+") || strings.Contains(topicName, "/#") {

		if strings.Contains(topicName, "#/") {
			return
		}

		for k, v := range maskSubscribers[topicName] {
			if v.id == client.id {
				maskSubscribers[topicName][k] = maskSubscribers[topicName][len(maskSubscribers)-1]
				maskSubscribers[topicName][len(maskSubscribers)-1] = maskSubscribers[topicName][k]
				maskSubscribers[topicName] = maskSubscribers[topicName][:len(maskSubscribers)-1]
				break
			}
		}

	}

	splitedPath := strings.Split(topicName, "/")[1:]

	if topicManager.roots[splitedPath[0]] == nil {
		return
	}

	node := topicManager.roots[splitedPath[0]]
	fullPath := "/" + topicManager.roots[splitedPath[0]].name

Loop:
	for key, name := range splitedPath[1:] {

		switch name {
		case "+":
			{
				for _, topic := range node.nodes {
					topicManager.Unsubscribe(client, strings.TrimRight(fullPath+"/"+topic.name+"/"+strings.Join(splitedPath[key+2:], "/"), "/"))
				}
				break Loop
			}
		case "#":
			{
				for _, topic := range node.nodes {
					topicManager.RecursiveDelSub(client, topic, fullPath+"/"+topic.name)
				}
				break Loop
			}
		default:
			{

				if node.nodes[name] == nil {
					continue Loop
				}

				fullPath += "/" + name
				node = node.nodes[name]

				if key == len(splitedPath)-2 {
					topicManager.DelSub(client, node, fullPath)
				}
			}
		}
	}
}

func (topicManager *TopicManager) RecursiveDelSub(client *Client, node *Node, fullPath string) {

	if client.subscriptions[fullPath] != nil {
		client.subscriptions[fullPath] = nil
		client.lock.Lock()
		delete(client.subscriptions, fullPath)
		client.lock.Unlock()

		for k, clientID := range node.subscribers {
			if client.id == clientID {
				node.subscribers[k] = node.subscribers[len(node.subscribers)-1]
				node.subscribers[len(node.subscribers)-1] = node.subscribers[k]
				node.subscribers = node.subscribers[:len(node.subscribers)-1]
				break
			}
		}
	} else {
		for _, topic := range node.nodes {
			topicManager.RecursiveDelSub(client, topic, fullPath+"/"+topic.name)
		}
	}
}

func (topicManager *TopicManager) DelSub(client *Client, topic *Node, fullPath string) {

	if client.subscriptions[fullPath] != nil {
		client.subscriptions[fullPath] = nil
		client.lock.Lock()
		delete(client.subscriptions, fullPath)
		client.lock.Unlock()

		for k, clientID := range topic.subscribers {
			if client.id == clientID {
				topic.subscribers[k] = topic.subscribers[len(topic.subscribers)-1]
				topic.subscribers[len(topic.subscribers)-1] = topic.subscribers[k]
				topic.subscribers = topic.subscribers[:len(topic.subscribers)-1]
				break
			}
		}
	}
}

func (topicManager *TopicManager) ClearUnusedTopics() {
	fmt.Println("TODO ClearUnusedTopics()")
}
