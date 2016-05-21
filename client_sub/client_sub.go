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


   A simple http client that simulates a subscriber.
   It registers for a topic and then polls (using http GET) for new messages on that topic

   To run:

   $ ./client_sub -topic=jobs -name=sub1 -poll=50 -port=6000

   By default it polls on port 3000 at 500 ms interval

*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type PubMessage struct {
	Message   string
	Published time.Time
}

func main() {
	var topic_name, subsciber_name string
	var poll_interval uint64
	var port int
	var ip string

	flag.StringVar(&topic_name, "topic", "sample_topic", "specifies the topic to subscribe to")
	flag.StringVar(&subsciber_name, "name", "sample_name", "specifies the subscriber name")
	flag.Uint64Var(&poll_interval, "poll", 500, "poll interval in milliseconds")
	flag.IntVar(&port, "port", 3000, "server port to subscribe to")
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip to subscribe to")

	flag.Parse()

	// register the client
	t := &http.Transport{}
	client := &http.Client{Transport: t}

	URL := fmt.Sprintf("http://%s:%d/%s/%s", ip, port, topic_name, subsciber_name)

	req, _ := http.NewRequest("POST", URL, nil)

	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != http.StatusCreated {
		fmt.Println("Error registering the subscriber to", topic_name)
		return
	}

	// start polling for new messages at the poll interval
	tick := time.Tick(time.Millisecond * time.Duration(poll_interval))

	for {

		<-tick

		resp, err = client.Get(URL)

		if err != nil {
			fmt.Println("Error in get", err)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			// fmt.Println("Topic Not Found")
		} else if resp.StatusCode == http.StatusNoContent {
			// fmt.Println("No New Messages")
		} else if resp.StatusCode == http.StatusOK {
			decoder := json.NewDecoder(resp.Body)
			var req PubMessage

			if err := decoder.Decode(&req); err != nil {
				fmt.Println("Error decoding the json body:", err)
			} else {
				fmt.Println("Message:", req.Message, ", Timestamp:", req.Published)
			}

		}

		resp.Body.Close()
	}

}
