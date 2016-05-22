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


   An HTTP based PubSub server

   $ pubsub -h
   Usage of pubsub:
   -ip string
    	ip address (default "127.0.0.1")
   -port int
    	server port to listen on (default 3000)

    Example:
    $ pubsub -port=6000

*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	// "github.com/nakdesai/pub-sub/pubsub"
	pubsub "github.com/nakdesai/pub-sub/pubsubScalable"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Maximum number of outstanding messages not pulled by the subscriber
const MAX_OUTSTANDING_MESSAGES int = 50

type PubSubInterface interface {
	Subscribe(topicName, subscriberName string)
	UnSubscribe(topicName, subscriberName string)
	Publish(topicName string, msg *pubsub.PubMessage)
	Get(topicName, subscriberName string) (*pubsub.PubMessage, error)
}

var (
	pb PubSubInterface
)

// publish a message on a topic
func publish(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// validate that the body of the POST request is json
	if r.Header["Content-Type"][0] != "application/json" {
		log.Println("Content-Type is not JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req pubsub.PubMessage

	if err := decoder.Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Published = time.Now()
	pb.Publish(params.ByName("topic_name"), &req)
	w.WriteHeader(http.StatusNoContent)
	return
}

// subscribe to a topic
func subscribe(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	pb.Subscribe(params.ByName("topic_name"), params.ByName("subscriber_name"))
	w.WriteHeader(http.StatusCreated)
	return
}

// Unsubscribe to a topic
func unsubscribe(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	pb.UnSubscribe(params.ByName("topic_name"), params.ByName("subscriber_name"))
	w.WriteHeader(http.StatusNoContent)
	return
}

// Pull a message for a topic
func getMsg(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	msg, err := pb.Get(params.ByName("topic_name"), params.ByName("subscriber_name"))

	// set the content-type to json
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		if err == pubsub.ErrSubNotFound || err == pubsub.ErrTopicNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else if err == pubsub.ErrNoNewMessages {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(*msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	return
}

func main() {

	var port int
	var ip string

	flag.IntVar(&port, "port", 3000, "server port to listen on")
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip address")

	flag.Parse()

	pb = pubsub.NewPubSub(MAX_OUTSTANDING_MESSAGES)

	router := httprouter.New()
	router.POST("/:topic_name", publish)
	router.POST("/:topic_name/:subscriber_name", subscribe)
	router.DELETE("/:topic_name/:subscriber_name", unsubscribe)
	router.GET("/:topic_name/:subscriber_name", getMsg)

	addr := fmt.Sprintf("%s:%d", ip, port)
	log.Fatal(http.ListenAndServe(addr, router))

}
