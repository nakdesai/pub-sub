/*
   Copyright (C) 2016 Nakul Desai

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.

   This file contains unit tests for the pubsub package.

   cmd to execute: "go test"

*/

package pubsub

import (
	"testing"
	"time"
)

// test Subscribe()
func TestSubscribe(t *testing.T) {
	NewPubSub(20)

	pb.Subscribe("testTopic", "testSubscriber")

	// It is necessary to block here on a channel so that it can yield the cpu and
	// let the event handler goroutine take over and perform the action
	<-time.After(time.Millisecond * 10)

	subMap, found := pb.topicMap["testTopic"]

	if !found {
		t.Errorf("Error inserting topic in map")
	}

	_, found = subMap["testSubscriber"]

	if !found {
		t.Errorf("Error inserting subscriber to topic map")
	}

	pb.Close()
}

// test Unsubscribe()
func TestUnsubscribe(t *testing.T) {
	NewPubSub(20)
	pb.Subscribe("testTopic", "testSubscriber")

	<-time.After(time.Millisecond * 10)

	subMap, found := pb.topicMap["testTopic"]

	if !found {
		t.Errorf("Error inserting topic in map")
	}

	_, found = subMap["testSubscriber"]

	if !found {
		t.Errorf("Error inserting subscriber to topic map")
	}

	pb.UnSubscribe("testTopic", "testSubscriber")

	<-time.After(time.Millisecond * 10)

	subMap, found = pb.topicMap["testTopic"]

	if !found {
		t.Errorf("Unsubscribe deleted topic !!")
	}

	_, found = subMap["testSubscriber"]

	if found {
		t.Errorf("Error unsubscribing from subscriber map")
	}

	pb.Close()
}

// test Publish()
func TestPublish(t *testing.T) {
	NewPubSub(20)
	pb.Subscribe("newTopic", "sub1")
	pb.Subscribe("newTopic", "sub2")
	pb.Subscribe("newTopic", "sub3")

	<-time.After(time.Millisecond * 10)

	pubMsg := &PubMessage{"msg", time.Now()}

	pb.Publish("newTopic", pubMsg)

	<-time.After(time.Millisecond * 10)

	recvMsg := <-pb.topicMap["newTopic"]["sub1"]

	if recvMsg.(*PubMessage).Message != "msg" {
		t.Errorf("Message not published to sub1")
	}

	recvMsg = <-pb.topicMap["newTopic"]["sub2"]

	if recvMsg.(*PubMessage).Message != "msg" {
		t.Errorf("Message not published to sub2")
	}

	recvMsg = <-pb.topicMap["newTopic"]["sub3"]

	if recvMsg.(*PubMessage).Message != "msg" {
		t.Errorf("Message not published to sub3")
	}

	// Ensure new subscriber only gets the new message
	pb.Subscribe("newTopic", "sub4")

	<-time.After(time.Millisecond * 10)

	pubMsg = &PubMessage{"newMsg", time.Now()}

	pb.Publish("newTopic", pubMsg)

	<-time.After(time.Millisecond * 10)

	recvMsg = <-pb.topicMap["newTopic"]["sub4"]

	if recvMsg.(*PubMessage).Message != "newMsg" {
		t.Errorf("New Message not published to sub4")
	}

	pb.Close()
}

// test Get()
func TestGetMsg(t *testing.T) {
	NewPubSub(20)
	pb.Subscribe("newTopic", "sub10")

	<-time.After(time.Millisecond * 10)

	pubMsg := &PubMessage{"sample_msg", time.Now()}

	pb.Publish("newTopic", pubMsg)

	<-time.After(time.Millisecond * 10)

	recvMsg, err := pb.Get("newTopic", "sub10")

	if err != nil {
		t.Errorf("Error recieving published message %s\n", err)
	}

	if recvMsg.Message != "sample_msg" {
		t.Errorf("Incorrect message received")
	}

	recvMsg, err = pb.Get("newTopic", "sub10")

	if err != ErrNoNewMessages {
		t.Errorf("subscriber buffer has more messages when it should be empty")
	}

	recvMsg, err = pb.Get("newTopic", "sub12")

	<-time.After(time.Millisecond * 10)

	if err != ErrSubNotFound {
		t.Errorf("topic has wrong subscriber")
	}

	recvMsg, err = pb.Get("newTopic12", "sub10")

	<-time.After(time.Millisecond * 10)

	if err != ErrTopicNotFound {
		t.Errorf("Invalid topic not flagged")
	}

	pb.Close()
}
