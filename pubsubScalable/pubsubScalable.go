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


   The pubsubScalable package provides a library for a simple pub-sub mechanism to subscribe to
   topics as well as publish messages to and pull messages from topics. The library spawns a fixed
   number of goroutines (topic managers) and and each goroutine is responsible to manager a part of
   the topic key space (topic name is hashed to determine the topic manager)

*/

package pubsubScalable

import (
	"errors"
	"hash/fnv"
	"runtime"
	"time"
)

// publisher message struct
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

// topic manager
type topicHandler struct {
	topicMap               map[string]map[string]chan interface{}
	maxOutStandingMessages int
	eventQueue             chan *request
}

type PubSub struct {
	topicHandlerChannelLst []chan *request
}

type req int

var (
	pb               *PubSub
	maxWorkers       uint32
	ErrSubNotFound   = errors.New("Subscriber Not Found")
	ErrTopicNotFound = errors.New("Topic Not Found")
	ErrNoNewMessages = errors.New("No New Messages for Subscriber")
)

const (
	ADD_SUB req = iota
	DEL_SUB
	GET_MSG
	POST_MSG
	CLOSE_QUEUE
)

// the length of the request queue
const REQUEST_QUEUE_SIZE int = 100

// Instantiate a new PubSub
func NewPubSub(maxOutStandingMsgs int) *PubSub {
	maxWorkers = uint32(runtime.NumCPU() - 1)

	pb = &PubSub{
		topicHandlerChannelLst: make([]chan *request, maxWorkers),
	}

	for i := 0; i < int(maxWorkers); i++ {
		w := newTopicHandler(maxOutStandingMsgs)
		pb.topicHandlerChannelLst[i] = w.eventQueue
	}

	return pb
}

// Instantiate a new topic handler
func newTopicHandler(maxOutStandingMessages int) *topicHandler {
	th := &topicHandler{
		maxOutStandingMessages: maxOutStandingMessages,
		eventQueue:             make(chan *request, REQUEST_QUEUE_SIZE),
	}

	go th.run()
	return th
}

// event handler goroutine to pull events from the event queue (channel) and process them
func (th *topicHandler) run() {
	th.topicMap = make(map[string]map[string]chan interface{})

	// on cleanup close the susbcriber goroutines
	defer func() {
		for topic := range th.topicMap {
			for sub := range th.topicMap[topic] {
				close(th.topicMap[topic][sub])
				delete(th.topicMap[topic], sub)
			}

			delete(th.topicMap, topic)
		}
	}()

	// event loop
	for r := range th.eventQueue {
		switch r.action {

		case ADD_SUB:
			ch := make(chan interface{}, th.maxOutStandingMessages)

			if _, found := th.topicMap[r.key]; !found {
				th.topicMap[r.key] = make(map[string]chan interface{})
			}

			th.topicMap[r.key][r.value.(string)] = ch

		case DEL_SUB:
			if subMap, found := th.topicMap[r.key]; found {
				delete(subMap, r.value.(string))
			}

		case POST_MSG:
			for _, ch := range th.topicMap[r.key] {
				select {
				case ch <- r.value:
				default:
				}
			}

		case GET_MSG:
			if subMap, found := th.topicMap[r.key]; found {
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
			close(th.eventQueue)
		}
	}
}

func getHashIdx(topic string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(topic))
	return h.Sum32() % maxWorkers
}

// subscribe to topics
func (pb *PubSub) Subscribe(topicName, subscriberName string) {
	pb.topicHandlerChannelLst[getHashIdx(topicName)] <- &request{ADD_SUB, topicName, subscriberName, nil}
}

// Unsubscribe to topics
func (pb *PubSub) UnSubscribe(topicName, subscriberName string) {
	pb.topicHandlerChannelLst[getHashIdx(topicName)] <- &request{DEL_SUB, topicName, subscriberName, nil}
}

// publish a message to a topic
func (pb *PubSub) Publish(topicName string, msg *PubMessage) {
	pb.topicHandlerChannelLst[getHashIdx(topicName)] <- &request{POST_MSG, topicName, msg, nil}
}

// pull the next message for the topic
func (pb *PubSub) Get(topicName, subscriberName string) (*PubMessage, error) {
	resp := make(chan response)

	pb.topicHandlerChannelLst[getHashIdx(topicName)] <- &request{GET_MSG, topicName, subscriberName, resp}

	r := <-resp

	if r.err != nil {
		return nil, r.err
	}

	return r.value.(*PubMessage), r.err
}

// close the pubsub
func (pb *PubSub) Close() {
	for _, ch := range pb.topicHandlerChannelLst {
		ch <- &request{action: CLOSE_QUEUE}
	}
}
