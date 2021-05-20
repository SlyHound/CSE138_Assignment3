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
The result of the error was because I was attempting to send curl requests on the wrong IP address of `10.10.0.4` with port number `8085`, which wasn't possible. After fixing that issue, Oleksiy replaced the IP address with `localhost` and kept the port number the same. The result made the curl requests go through correctly. 

## Helped by professor Lindsey Kuper during her office hours.
- Oleksiy provided a Lamport diagram describing his thought process of how to ensure that replica's communicated effectively regarding whether or not a replica is currently down. Professor Kuper indicated that Oleksiy's approach makes sense and that broadcasting requests, even when the work seems redundant was a way to ensure even when a replica is down, every replica is eventually able to receive a request. 

## Tutored by Aria Diamond.
- Aria helped Oleksiy debug a race condition that was occurring as a result of a write and read concurrent access. Aria suggested using either a global mutex or passing them separately to each function as having a global mutex is bad practice. The change suggested by Aria helped to remove the race condition.

## Discussing with Vinay Venkat.
Oleksiy had a question regarding how to get data from another replica once the current replica comes back up. Vinay suggested using an additional endpoint in which a GET request could be administered to retrieve the local key-value store following the prior PUT requests. The suggestion of using another endpoint was implemented and is present in which the GET request to retrieve the key-value store is broadcasted after the PUT requests. 

# Citations
- The package [http](https://golang.org/pkg/net/http/) was used to ensure that messages could be forwarded using the `http.Do()` function. Also, recently Oleksiy learned that the package [Gin](https://github.com/gin-gonic/gin) (which our group has been using since assignment 2 to listen and serve requests), when running the `router.Run()` command, it actually runs `http.ListenAndServe()` in its implementation! Oh also, in the http package library, a new request had to be built using `http.NewRequest()` to be able to send messages from one replica to another.
- To ensure that the server could serve requests and send them (in the case of view requests), the [sync](https://golang.org/pkg/sync/) library was utilized to be to spring up multiple threads without issues. Patrick actually suggested using mutexes from this library. The `Lock()` and `Unlock()` functions were used from library. Those two functions helped to ensure that there aren't any concurrent writes and reads happening at the same time. The `RWMutex` could be used instead to allow concurrent reads and only a single writer. As described by Aria and Patrick, these types of locks can help to speed up the program, although they weren't used since the only locks Oleksiy is familiar with is the ones previously described. Piggybacking off of synchronization, Zach suggested using the flag [--race](https://golang.org/doc/articles/race_detector) to check for race conditions upon testing the code.
- Lecture 5, Vector Clocks Algorithm (with a twist): The algorithm described here was the basis of our causal consistency implementation. We used the vector clock from this as the causal-metadata from replica to replica to keep PUT and DELETE requests consistent. Just like the algorithm, we count sends as events, but we do not count deliver as events.

# Team Contributions

Oleksiy Omelchenko: View Operations
- Sending periodic health checks (pings) to every replica in order to keep track of which replicas are up or down
- GET, DELETE, and PUT view operations

Zach Zulanas: Broadcasts
- Sending / forwarding PUT and DELETE requests recieved from client to ensure consistency between replicas
- Implemented replica-to-replica endpoint to allow replicas to send changes to one another

Jackie Garcia: Causal Consistency
- Implemented Vector Clocks for every replica and send as causal-metadata (described in mechanism-description)
- Wrote Causal Broadcast algorithm function canDeliver() that compares send and recieve VCs to determine if a message is deliverable
- Wrote causal consistency implementation explanation in mechanism-description
