# Ensuring Causal Consistency

## causal-metadata
For our causal-metadata, we pass in an array of length 4: [R1, R2, R3, P]
- [R1, R2, R3] represent the Vector Clock values for each of our three replicas
- P indicates what the position of the sender replica is
  -  For example, if Replica 1 was sending its VC [2, 0, 2], causal-metadata would be [2, 0, 2, 0]

## Causal Broadcasts
In order to keep consistent replicas, whichever process serves a request from the client (that updates the key-value store) sends this update to the other replicas. These are PUT and DELETE requests.

We implemented broadcast as a series of unicast messages:
- The sender replica will increment its own position in its VC (since a broadcast is one event) and then send the message to the other replicas
- In a new thread, the sender will wait for a response from the replica
  - If it recieves a success message, this thread exits
  - If it recieves an error, it will send the message again
    - It will continue sending messages until it recieves a success message

## canDeliver() Algorithm
In order to determine whether a replica can deliver the message it has just recieved, we use the Causal Broadcast Algorithm from Lecture 5, pseudocode below

```
canDeliver(sender vector clock (senderVC), reciever vector clock (thisVC)):

  if the conditions are met, return true:
    senderVC[sender slot] = thisVC[sender slot] + 1
    senderVC[every not sender slot] <= thisVC[every not sender slot]
    
  else return false
```

## Causal Delivery
When a replica recieves a message, it uses the `canDeliver()` conditions to determine if delivering this message would violate causal consistency.
- If it is safe, the replica updates its vector clock: max (sender, current) for all values,  increments its own (since we are counting delivery as an event), and delivers the message
  - It then returns this updated VC and a success message to the sender replica
- If, however, the message is not safe to be delivered, the replica will return an error message to the sender
  - If it was unsuccessful, this means it is still waiting on previous messages that it must deliver before this one will be causally consistent
  - The sender will know that the message could not be delivered, and it will send again
  - This process will continue until the reciever's VC has been updated and it can deliver this message
