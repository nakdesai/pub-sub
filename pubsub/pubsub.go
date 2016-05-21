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


   The pubsub package provides a library for a simple pub-sub mechanism to subscribe to
   topics as well as publish messages to and pull messages from topics.

*/

package pubsub

import (
	"errors"
	"time"
)

// publsiher message struct
type PubMessage struct {
	Message   string
	Published time.Time
}

// request event struct
type request struct {
	action req
	key    string
	value  interface{}
	result chan response
}

// response event struct
type response struct {
	value interface{}
	err   error
}

type PubSub struct {
	topicMap               map[string]map[string]chan interface{}
	maxOutStandingMessages int
	reqCh                  chan *request
}

var (
	pb               *PubSub
	ErrSubNotFound   = errors.New("Subscriber Not Found")
	ErrTopicNotFound = errors.New("Topic Not Found")
	ErrNoNewMessages = errors.New("No New Messages for Subscriber")
)

type req int

// event type
const (
	ADD_SUB req = iota
	DEL_SUB
	GET_MSG
	POST_MSG
	CLOSE_QUEUE
)

// the length of the request queue
const REQUEST_QUEUE_SIZE int = 200

// Instantiate a new PubSub
func NewPubSub(maxOutStandingMsgs int) *PubSub {
	pb = &PubSub{
		maxOutStandingMessages: maxOutStandingMsgs,
		reqCh: make(chan *request, REQUEST_QUEUE_SIZE),
	}

	go pb.run()
	return pb
}

// event handler goroutine to pull events from the request queue (channel) and process them
func (pb *PubSub) run() {
	pb.topicMap = make(map[string]map[string]chan interface{})

	defer func() {
		for topic := range pb.topicMap {
			for sub := range pb.topicMap[topic] {
				close(pb.topicMap[topic][sub])
				delete(pb.topicMap[topic], sub)
			}

			delete(pb.topicMap, topic)
		}

		pb.topicMap = nil

	}()

	// event loop
	for r := range pb.reqCh {
		switch r.action {

		case ADD_SUB:
			ch := make(chan interface{}, pb.maxOutStandingMessages)

			if _, found := pb.topicMap[r.key]; !found {
				pb.topicMap[r.key] = make(map[string]chan interface{})
			}

			pb.topicMap[r.key][r.value.(string)] = ch

		case DEL_SUB:
			if subMap, found := pb.topicMap[r.key]; found {
				delete(subMap, r.value.(string))
			}

		case POST_MSG:
			for _, ch := range pb.topicMap[r.key] {
				select {
				case ch <- r.value:
				default:
				}
			}

		case GET_MSG:
			if subMap, found := pb.topicMap[r.key]; found {
				if subCh, found := subMap[r.value.(string)]; found {
					select {
					case v := <-subCh:
						r.result <- response{v, nil}
					default:
						r.result <- response{nil, ErrNoNewMessages}
					}
				} else {
					r.result <- response{nil, ErrSubNotFound}
				}
			} else {
				r.result <- response{nil, ErrTopicNotFound}
			}

		case CLOSE_QUEUE:
			close(pb.reqCh)
		}
	}
}

// subscribe to topics
func (pb *PubSub) Subscribe(topicName, subscriberName string) {
	pb.reqCh <- &request{ADD_SUB, topicName, subscriberName, nil}
}

// Unsubscribe to topics
func (pb *PubSub) UnSubscribe(topicName, subscriberName string) {
	pb.reqCh <- &request{DEL_SUB, topicName, subscriberName, nil}
}

// publish a message to a topic
func (pb *PubSub) Publish(topicName string, msg *PubMessage) {
	pb.reqCh <- &request{POST_MSG, topicName, msg, nil}
}

// pull the next message for the topic
func (pb *PubSub) Get(topicName, subscriberName string) (*PubMessage, error) {
	resp := make(chan response)

	pb.reqCh <- &request{GET_MSG, topicName, subscriberName, resp}

	r := <-resp

	if r.err != nil {
		return nil, r.err
	}

	return r.value.(*PubMessage), r.err
}

// close the pubsub
func (pb *PubSub) Close() {
	pb.reqCh <- &request{action: CLOSE_QUEUE}
}
