# Acknowledgments

`Oleksiy Omelchenko:`
## He was helped by Patrick Redmond on four different ocassions. 
- The first of which was understanding how to handle replica's knowing when a server goes up or goes down for view operations. The suggestion to handle replica's is to keep a sort of queue data structure that kept track of which replica's were currently up by sending messages to each other periodicially.
- The second situation in which he received help by Patrick is to perform all GET requests and then all DELETE requests for view operations. Previously a GET request was followed by a DELETE request if a replica was found to be currently down. His suggestion was taken into consideration and applied to the view operations with resounding success.
- The third time was during section where Patrick drew out some casual consistency Lamport diagrams and described the vector clocks on each step. The Lamport diagrams helped the team better understand casual consistency through the comments listed on the images. 
- The final time in which Oleksiy received help was when he was having trouble running the testing script that resulted in the following error: 
```
ConnectionRefusedError: [WinError 10061] No connection could be made because the target machine actively refused it
``` 
The result of the error was because Oleksiy was attempting to send curl requests on the wrong IP address of `10.10.0.4` with port number `8085`, which wasn't possible. After fixing that issue, Oleksiy replaced the IP address with `localhost` and kept the port number the same. The result made the curl requests go through correctly. 

## Helped by professor Lindsey Kuper during her office hours.
- Oleksiy provided a Lamport diagram describing his thought process of how to ensure that replica's communicated effectively regarding whether or not a replica is currently down. Professor Kuper indicated that Oleksiy's approach makes sense and that broadcasting requests, even when the work seems redundant was a way to ensure even when a replica is down, every replica is eventually able to receive a request. 

## Tutored by Aria Diamond.
- Aria helped Oleksiy debug a race condition that was occurring as a result of a write and read concurrent access. Aria suggested using either a global mutex or passing them separately to each function as having a global mutex is bad practice. The change suggested by Aria helped to remove the race condition.

## Discussing with Vinay Venkat.
Oleksiy had a question regarding how to get data from another replica once the current replica comes back up. Vinay suggested using an additional endpoint in which a GET request could be administered to retrieve the local key-value store following the prior PUT requests. The suggestion of using another endpoint was implemented and is present in which the GET request to retrieve the key-value store is broadcasted after the PUT requests. 

`Zachary Zulanas:`

## Matthew Boisvert
* Helped with getting the idea of how to set up causal consistency and where to lock our variables
* Gave ideas on how to avoid deadlock and maintain consistency

## Arjun Loganathan
* Helped Oleksiy and I understand how to implement causal consistency and helped debug our code when we kept running into errors. We implemented his ideas/algorithm for our causal delivery and causal consistency where we rebroadcast the entire casual history of our replicas to ones that start back up again/come online.

# Citations
- The package [http](https://golang.org/pkg/net/http/) was used to ensure that messages could be forwarded using the `http.Do()` function. Also, recently Oleksiy learned that the package [Gin](https://github.com/gin-gonic/gin) (which our group has been using since assignment 2 to listen and serve requests), when running the `router.Run()` command, it actually runs `http.ListenAndServe()` in its implementation! The function `http.ListenAndServe()` is the default way in which Go is able to listen on a socket for any incoming requests that the replica must serve. Oh also, in the http package library, a new request had to be built using `http.NewRequest()` to be able to send messages from one replica to another. The `http.NewRequest()` was used to ensure that we have replica to replica communication for broadcasting view requests as well as key-value store requests for casual consistency. 
- To ensure that the server could serve requests and send them (in the case of view requests), the [sync](https://golang.org/pkg/sync/) library was utilized to be able to spring up multiple threads without issue. The server had to be able to listen for incoming requests by the client or replica alike and make sure that if a replica went down or came back up, that all other replica's knew about it. To be able to get another thread up and running, Oleksiy used Go's `go` syntax as suggested by this [guide](https://golangbyexample.com/goroutines-golang/). There was notably two different threads running, the health check thread meant for checking if a replica was up or down and the main thread for serving requests and delivering responses. By creating a seperate thread, Oleksiy learned that creating threads was much easier than what he was used to with his prior experience with using threads in C for CSE130. In regards to ensuring mutual exclusion on reads and writes, Patrick suggested using mutexes from the sync library. The `Lock()` and `Unlock()` functions were used from that library. The `Lock()` and `Unlock()` functions helped to ensure that there wasn't any concurrent writes and reads happening at the same time. Oleksiy also learned from both Aria and Patrick that Go has what're called the `RWMutex` that could be used instead to allow concurrent reads and only a single writer. Previously our group only knew of the default mutually exclusive locks and never about the existence of read and write locks as mentioned by Patrick during office hours.
- Although, not used in the later versions of the code, [channels](https://golangbot.com/channels/) were looked into for ease of access to be able to transmit views easily between threads. In the struggle of learning about channels, the group learned that a send channel by default blocks until some other thread receives from that channel. The unfortunate situation was that we wanted to utilize this functionality on the sender and receiver side of a replica. However, if we placed a send channel before sending a request, the request would never send as a result of the receive channel being located in the delivery endpoint. If we did the opposite, as in send into the channel at the delivery endpoint, then the response would have never been sent and the receive channel line wouldn't have been reached. In all, the channels were originally meant to be used to communicate views and key-value store information between replica's.   
- Now onto synchronization techniques, Zach suggested using the flag [--race](https://golang.org/doc/articles/race_detector) to check for race conditions upon testing the code. Using that race flag in the Dockerfile helped in debugging to know where locks were forgotten to be placed. By using the flag, we accomplished being able to find all lines at which write and read race conditions occur provided by the stack trace outputted emitted by the flag. 
- Lecture 5, Vector Clocks Algorithm (with a twist): The algorithm described here was the basis of our causal consistency implementation. We used the vector clock from this as the causal-metadata from replica to replica to keep PUT and DELETE requests consistent. Just like the algorithm, we count sends as events, but we do not count deliver as events. Using the vector clock algorithm, we gained knowledge how vector clocks are used to enforce casual consistency. The drawback of attempting to implement the algorithm was that there was some confusion that ensued attempting to debug a given replica's vector clock when dozens of requests occur in the midst of it all.
- Assignment 2 code was used as a skeleton to build off of for delete, get, and put requests to the key-value store endpoint. Although, the groundwork was laid, additional endpoints had to be added for when replica's wished to replicate their key-value stores and views as the clients wouldn't be calling those endpoints. In the beginning it was confusing as to how we would be able to communicate between replica's without using the already existing endpoints. However, Patrick and Vinay taught us that additional endpoints could be utilized without any negative consequences as we would use them for our personal use and as such can be structured in any form we see fit.   

# Team Contributions

Oleksiy Omelchenko: View Operations
- Sending periodic health checks (pings) to every replica in order to keep track of which replicas are up or down (more in-detail provided in the mechanism-description document).
- The operations for the view operations include broadcasting GET requests on a 1 second interval to check that a replica is up or down. If a replica is down, then broadcast a DELETE request and finally broadcast a PUT request from the replica that comes up followed by a broadcasted GET to a custom endpoint. At the custom endpoint, a replica returns its key value store back to the brought up replica. 

Zach Zulanas: Broadcasts
- Sending / forwarding PUT and DELETE requests recieved from client to ensure consistency between replicas
- Implemented replica-to-replica endpoint to allow replicas to send changes to one another
- Built out Causal consistency distribution algorithm with Jackie and Oleksiy
- Created PutRequest(), ReplicatePut(), DeleteRequest(), ReplicateDelete(), GetRequest(), and setupRouter() functions. All deal with building routes/endpoints for client <-> server communication and replica <-> replica communication

Jackie Garcia: Causal Consistency
- Implemented Vector Clocks for every replica as causal-metadata (described in the mechanism-description).
- updateVC() function to take the pointwise maximum of a client message VC and current replica's VC.
- Wrote Causal Broadcast algorithm function canDeliver() that compares send and receive VCs to determine if a message is deliverable.
- Maintain and update vector clocks before sending causal broadcasts to other replicas.
- Wrote causal consistency implementation explanation in mechanism-description.

In addition to this, we peer programmed and debugged the implementation of our causal consistency VCs, and how they get sent across the replica specific endpoint.
