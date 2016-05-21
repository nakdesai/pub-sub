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


  This file contains unit tests for main.go

*/
package main

import (
	"net/http"
	"net/http/httptest"
	"pubsub"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

// mock the PubSub type by implementing the PubSubInterface interface
type mockPB struct{}

func (m *mockPB) Subscribe(topicName, subscriberName string) {
	return
}

func (m *mockPB) UnSubscribe(topicName, subscriberName string) {
	return
}

func (m *mockPB) Publish(topicName string, msg *pubsub.PubMessage) {
	return
}

func (m *mockPB) Get(topicName, subscriberName string) (*pubsub.PubMessage, error) {
	return &pubsub.PubMessage{Message: "msg", Published: time.Now()}, nil
}

// test subscribe
func TestSubscribe(t *testing.T) {
	pb = &mockPB{}

	req, _ := http.NewRequest("POST", "http://localhost:3000/topic/message", nil)

	params := []httprouter.Param{
		{
			Key:   "topic_name",
			Value: "topic1",
		},
		{
			Key:   "subscriber_name",
			Value: "sub1",
		},
	}
	w := httptest.NewRecorder()
	subscribe(w, req, params)

	if w.Code != http.StatusCreated {
		t.Errorf("Incorrect http status code for subscribe operation")
	}

}

// test unsubscribe
func TestUnSubscribe(t *testing.T) {
	pb = &mockPB{}

	req, _ := http.NewRequest("DELETE", "http://localhost:3000/topic/message", nil)

	params := []httprouter.Param{
		{
			Key:   "topic_name",
			Value: "topic1",
		},
		{
			Key:   "subscriber_name",
			Value: "sub1",
		},
	}
	w := httptest.NewRecorder()
	unsubscribe(w, req, params)

	if w.Code != http.StatusNoContent {
		t.Errorf("Incorrect http status code for unsubscribe operation")
	}

}

// test getMsg
func TestGetMsg(t *testing.T) {
	pb = &mockPB{}

	req, _ := http.NewRequest("GET", "http://localhost:3000/topic/message", nil)

	params := []httprouter.Param{
		{
			Key:   "topic_name",
			Value: "topic1",
		},
		{
			Key:   "subscriber_name",
			Value: "sub1",
		},
	}
	w := httptest.NewRecorder()
	getMsg(w, req, params)

	if w.Code != http.StatusOK {
		t.Errorf("Incorrect http status code for getting next message from topic")
	}

}

// TODO
func TestPublish(t *testing.T) {
}
