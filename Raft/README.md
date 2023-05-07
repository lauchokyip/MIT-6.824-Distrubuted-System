# Lab 2A Q&As

1) What is applyCh use for?
as each Raft peer becomes aware that successive log entries are committed, the peer should send an ApplyMsg to the service. </br>

The whole process is as below (Reference: http://thesecretlivesofdata.com/raft/#replication):
-> First the User will send a change to the leader
-> The change is appended to the leader's log
-> The change is then sent to the peers on the next heartbeat
-> An entry is then committed **on the leader** once a majority of followers acknowledge it
-> Then the response is sent to the client (I guess this is the point of time where the leader will send ApplyMsg to the applyCh)
-> Then I guess the leader will send AppendEntries to the peers to commit ?

I noticed, the leader will send AppendEntries to peers to commit the messages after the leader send ApplyMsg to applyCh


2) When do I have to lock the data?
Good rule of thumb is </br>
> A more pragmatic approach starts with the observation that if there were no concurrency (no simultaneously executing goroutines), you would not need locks at all. But you have concurrency forced on you when the RPC system creates goroutines to execute RPC handlers, and because you need to send RPCs in separate goroutines to avoid waiting. You can effectively eliminate this concurrency by identifying all places where goroutines start (RPC handlers, background goroutines you create in Make(), &c), acquiring the lock at the very start of each goroutine, and only releasing the lock when that goroutine has completely finished and returns.
> 
> the next step is to find places where the code waits, and to add lock releases and re-acquires (and/or goroutine creation) as needed, being careful to re-establish assumptions after each re-acquire. You may find this process easier to get right than directly identifying sequences that must be locked for correctness.

3) How to check if a leader election should be started ? </br>
If election timeout elapses. Heartbeat always have to be shorter than election timeout

4) For RequestVote RPC , what does candidate's log is at least as up-to-date as receiver's log mean? </br>
> Raft determines which of two logs is more up-to-date by comparing the index and term of the last entries in the logs. If the logs have last entries with different terms, then the log with the later term is more up-to-date. If the logs end with the same term, then whichever log is longer is more up-to-date.

5) Why do we need to use election timeout larger than 150ms-300ms? </br>
Imagine if the heartbeat is sent every 100ms since the lab limits 10 heartbeat every 1s, so about 1 heartbeart every 100ms, the election will happen pretty frequently


6) What is the difference between `lastLogTerm` and `currentTerm` and when would their value be different? </br>
`currentTerm` will always be updated when receiving a RPC to indicate which node has the most up-to-date information
`lastLogTerm` is the term of the last log entry in a node's log. It is used during leader election to compare the logs of different nodes and determine which node has the most up-to-date log.

7) If two nodes become candidate and start elections, and one node receives the majority vote, and the other candidate receives the AppendEntries from the leader, how does it know if it's a valid leader?
> While waiting for votes, a candidate may receive an AppendEntries RPC from another server claiming to be leader. If the leader’s term (included in its RPC) is at least as large as the candidate’s current term, then the candidate recognizes the leader as legitimate and returns to follower state. If the term in the RPC is smaller than the candidate’s current term, then the candidate rejects the RPC and continues in candidate state.

In our case, AppendEntries will check for the condition `if args.Term < rf.currentTerm` this will make sure the leader’s term (included in its RPC) is at least as large as the candidate’s current term.

8) When would `votedFor` gets reset? </br>
Everytime time it has to step down, which is when
> If RPC request or response contains term T > currentTerm:
> set currentTerm = T, convert to follower

9) How to interpret this sentence? If election timeout elapses without receiving AppendEntries RPC from current leader or granting vote to candidate, convert to candidate </br>
In another way to phrase it, there are two ways the election timeout will be reset, which are receiving valid AppendEntries RPC **from current learder** or it has granted vote 

# Lab 2B Q&As
1) The hint states that we need to implement the election restriction (section 5.4.1 in the paper). What is the restriction ? </br>
> Raft uses a simpler approach where
> it **guarantees that all the committed entries from previous terms are present on each new leader from the moment of its election** Raft uses the voting process to prevent a candidate from winning an election unless its log contains all committed entries. A candidate must contact a majority of the cluster in order to be elected, which means that every committed entry must be present in at least one of those servers.

2) What does `Start` do ? 
It's called by any service that uses Raft, can be a k/v server and the service would want the command to be appended to Raft's log. 

3) The comment for `Start` says it should "start the agreement and return immediately" What does start the agreement mean?</br>
I think it means to append the command to the log entry 

4) After the `Start` function append the log to the leader, how does the leader know new log has been sent and start AppendEntries?
The leader will initialize `nextIndex` to 0 , and if log is appended to the leader, the last log index will be greater than `nextIndex` 

5) The paper mentions about retrying, does that mean if AppendEntries fails , the same go routine that runs it will keep resend the AppendEntries until it succeeds? Or it will return and wait until the next heartbeat timer elapses to fix it? </br>
It does not make sense for AppendEntries to try again in the same go routine because if it keeps failing, and the next goroutine is fired, they will be two go routines doing the same thing.

6) The paper mentions about If successful: update nextIndex and matchIndex for follower, what value should we update nextIndex and matchIndex to? </br>
This is tricky because it's not on the figure 2. In figure 5.3, it says 
> When a leader first comes to power, it initializes all nextIndex values to the index just after the last one in its log (11 in Figure 7)
so it's initialized to len(entries) if we assume the index of LogEntries start with 0

7) If we find the majority of matchIndex, do we want to stop waiting for other AppendEntries? 

8) How do we make sure when stepping down it's happening atomically before switching states?

9) When will be commit the messages?
in function `sendAppendEntriesToAllServers` and `AppendEntriesRPC` whenever `commitIndex` is changed

## Things to be aware of when doing distributed system programming
1) 
## Stupid mistakes made:
1) have a select statement, and a unbuffered channel
```go
func foo() {
    ch<- struct{}
}

select {
    case <-ch:
    case <-ch1:
}
```
If ch1 is passed through, the select statement will end, so `func` will be blocked forever

2) another select statement
```go
func foo() <-chan struct{}{
    ch := make(struct{}, 1)
    time.Sleep(10 * time.Seconds)

    return ch
}

select {
    case <-foo():
    case <-ch1:
}
```
There are some cases where select statement will select `foo()` and it will be blocked and other channel will not be executed

## Progress
May 5 - Trying to figure out ApplyMsg
May 7 - Passed 2B

## Good resources:
* http://nil.csail.mit.edu/6.824/2022/notes/raft_diagram.pdf
* https://www.cs.princeton.edu/courses/archive/fall18/cos418/docs/p7-raft-no-solutions.pdf
* https://courses.cs.duke.edu/fall21/compsci514/internal/slides/lecture17.pptx
* http://nil.csail.mit.edu/6.824/2022/notes/raft_diagram.pdf