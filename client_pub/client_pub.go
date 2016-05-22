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


    A simple http client that simulates a publisher publishing messages to 1 or more topics.
    Depending on the number of topics specified (default 1), the topics are called topic1, topic2, etc.

    At random intervals of time it will publish a message to that topic
    i.e message1
        message2
        message3
        ....

    To run :

    $ client_pub -port=6000 -topic=jobs -message=doctor -interval=800 -num=5

    By default it publishes to port 3000 and the default max interval between successive messages is 500 ms

*/
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type PubMessage struct {
	Message string
}

func main() {
	var topic_name, message string
	var max_interval int
	var port int
	var pubMsg PubMessage
	var ip string
	var num int

	flag.StringVar(&topic_name, "topic", "topic", "specifies the topic to post to")
	flag.StringVar(&message, "message", "message", "specifies the message to post to the topic")
	flag.IntVar(&max_interval, "interval", 500, "max interval in milliseconds between succesive posts")
	flag.IntVar(&port, "port", 3000, "server port to publish to")
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip to publish to")
	flag.IntVar(&num, "num", 1, "number of topics to publish to")

	flag.Parse()

	// initialize the signal handlers
	sigHandlerCh := make(chan os.Signal, 1)
	signal.Notify(sigHandlerCh, syscall.SIGINT, syscall.SIGTERM)

	// done channel to terminate the publisher goroutines
	done := make(chan struct{})

	go func() {
		<-sigHandlerCh
		fmt.Println("Terminating the Publishers ...")
		close(done)

	}()

	var wg sync.WaitGroup

	for index := 1; index <= num; index++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			rand.Seed(time.Now().Unix())

			t := &http.Transport{}
			client := &http.Client{Transport: t}

			URL := fmt.Sprintf("http://%s:%d/%s%d", ip, port, topic_name, id)

			i := 0

			for {
				select {
				case <-time.After(time.Millisecond * time.Duration(rand.Intn(max_interval))):

					i++
					pubMsg.Message = fmt.Sprintf("%s%d", message, i)

					b, err := json.Marshal(pubMsg)

					if err != nil {
						fmt.Println("Error marshaling json:", err)
						continue
					}

					req, _ := http.NewRequest("POST", URL, bytes.NewBuffer(b))

					req.Header.Set("Content-Type", "application/json")

					resp, err := client.Do(req)

					if err != nil {
						fmt.Println("Error in get", err)
						continue
					}

					if resp.StatusCode != http.StatusNoContent {
						fmt.Println("Error publishing message")
					}

					resp.Body.Close()

				case <-done:
					return
				}

			}

		}(index)
	}

	wg.Wait()
	fmt.Println("Client Terminated")
}
