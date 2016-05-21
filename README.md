# pub-sub
A simple http based pub-sub that allows clients to publish messages to topics, subscribe/unsubscribe to topics and poll for new messages for a topic.

# HTTP Endpoints
Subscribe (subscribe to topic topic_name with username subscriber_name):
    POST /{topic_name}/{subscriber_name}
    
    Response: 201

Unsubscribe (unsubscribe subscriber_name from topic topic_name:
    DELETE /{topic_name}/{subscriber_name} 
    
    Response: 204
    
Publish (publish message to topic topic_name)
    POST /{topic_name}
    
        {
            "message": <message string>
        }
        
    Response: 204

Get (get the next new message for topic topic_name for subscriber subscriber_name)
    GET /{topic_name}/{subscriber_name}
    
    Response:
        204 (No New Messages)
        404 (No Subscriber named subscriber_name or no topic named topic_name)
        200 OK
            {
                "message": <message string>,
                "published": <time stamp>
            }
# Install
  $ go get github.com/nakdesai/pub-sub
  
  publisher client
  
  $ go install github.com/nakdesai/pub-sub/client_pub
  
  subscriber client
  
  $ go install github.com/nakdesai/pub-sub/client_sub
            
# Usage
   $ pub-sub --help
   Usage of pub-sub:
   -ip string
    	 ip address (default "127.0.0.1")
   -port int
    	 server port to listen on (default 3000)

   Example: 
       pub-sub -port=6000 (will start server listening on port 6000 on localhost)
       
# Helper Clients
  These are helper clients that simluate publishers and subscribers
  
  publisher client:
      $ client_pub --help
      
      Usage of client_pub:
      -interval int
    	    max interval in milliseconds between succesive posts (default 500)
      -ip string
    	    ip to publish to (default "127.0.0.1")
      -message string
    	    specifies the message to post to the topic (default "sample_name")
      -port int
    	    server port to publish to (default 3000)
      -topic string
    	    specifies the topic to post to (default "sample_topic")

       Example:
       $ client_pub -port=6000 -topic=jobs -message=doctor -interval=500
       
       This will start publishing messages to localhost port 6000 for topic "jobs" at a max interval of 500 ms between successive messages.
       the message will be appended by a counter i.e in this case doctor1, doctor2, ..etc
       
  subscriber client:
      $ client_sub --help
      
      Usage of client_sub:
      -ip string
    	    ip to subscribe to (default "127.0.0.1")
      -name string
    	    specifies the subscriber name (default "sample_name")
      -poll uint
    	    poll interval in milliseconds (default 500)
      -port int
    	    server port to subscribe to (default 3000)
      -topic string
    	    specifies the topic to subscribe to (default "sample_topic")

      Example:
      $ client_sub -topic=jobs -name=sub1 -poll=500 -port=6000
      
      This will subscribe to topic "jobs" and poll for new messages at a poll interval of 500 ms
