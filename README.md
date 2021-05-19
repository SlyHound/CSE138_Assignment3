## Team Contributions

Oleksiy Omelchenko: View Operations
- Sending periodic health checks (pings) to every replica in order to keep track of which replicas are up or down
- GET, DELETE, and PUT view operations


Zach Zulanas: Broadcasts
- Sending / forwarding PUT and DELETE requests recieved from client to ensure consistency between replicas
- Multithreading implementation of send requests to replicas - keeps sending a PUT or DELETE until replica can deliver

Jackie Garcia: Causal Consistency
- Implemented Vector Clocks for every replica and send as causal-metadata
- Wrote Causal Broadcast algorithm function canDeliver() that compares send and recieve VCs to determine if a message is deliverable
