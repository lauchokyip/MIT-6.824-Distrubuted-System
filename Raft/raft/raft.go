package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"

	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type State int

const (
	Follower State = iota
	Candidate
	Leader
)

const (
	heartbeatInterval = 100 * time.Millisecond

	// choosing 800 and 1200 because for TestInitialElection2A
	// it will check if term is the same and you don't want to
	// be timeout too fast when the network packages are lost
	// when sending AppendEntries
	electionTimeoutMin = 800  // in ms
	electionTimeoutMax = 1200 // in ms
)

// A Go object implementing a single Raft peer.
type Raft struct {
	sync.Mutex                     // Lock to protect shared access to this peer's state
	peers      []*labrpc.ClientEnd // RPC end points of all peers
	persister  *Persister          // Object to hold this peer's persisted state
	me         int                 // this peer's index into peers[]
	dead       int32               // set by Kill()

	state State
	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	PersistentState
	VolatileState
	applyCh chan ApplyMsg

	// custom
	appendEntriesChan chan struct{}
	grantVoteChan     chan struct{}
	stepDownChan      chan struct{}
	readyToCommitChan chan struct{}
}

type LogEntry struct {
	Command interface{}
	Term    int
}

type PersistentState struct {
	// latest term server has seen (initialized to 0
	// 	on first boot, increases monotonically)
	currentTerm int
	// 	candidateId that received vote in current
	// term (or null if none)
	votedFor int
	// each entry contains command
	// for state machine, and term when entry
	// was received by leader (first index is 1)
	logs []LogEntry
}

type VolatileState struct {
	// index of highest log entry known to be
	// committed (initialized to 0, increases
	// monotonically)
	commitIndex int
	// 	index of highest log entry applied to state
	// machine (initialized to 0, increases
	// monotonically)
	lastApplied int

	// Below are only for leaders:

	// for each server, index of the next log entry
	// to send to that server (initialized to leader
	// last log index + 1)
	nextIndex []int
	// 	for each server, index of highest log entry
	// known to be replicated on server
	// (initialized to 0, increases monotonically)
	matchIndex []int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	// Your code here (2A).
	rf.Lock()
	defer rf.Unlock()
	log.Printf("server %d has term %d\n", rf.me, rf.currentTerm)

	return rf.currentTerm, rf.state == Leader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

// example AppendEntriesArgs RPC arguments structure.
// field names must start with capital letters!
type AppendEntriesArgs struct {
	// Your data here (2A, 2B).
	Term         int        // leader's term
	LeaderID     int        // so follower can redirect clients
	PrevLogIndex int        // index of log entry immediately preceding new ones
	PrevLogTerm  int        // term of prevLogIndex entry
	Entries      []LogEntry // log entries to store (empty for heartbeat; may send more than one for efficiency)
	LeaderCommit int        // leader’s commitIndex
}

// example AppendEntriesReply RPC reply structure.
// field names must start with capital letters!
type AppendEntriesReply struct {
	// Your data here (2A).
	Term   int  // currentTerm, for leader to update itself
	Sucess bool // true if follower contained entry matching prevLogIndex and prevLogTerm
}

type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int // candidate’s term
	CandidateId  int // candidate requesting vote
	LastLogIndex int // index of candidate’s last log entry
	LastLogTerm  int // term of candidate’s last log entry
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int  // currentTerm, for candidate to update itself
	VoteGranted bool // true means candidate received vote
}

// stepdown will set the term to the higher term and convert to follower
// it will reset votedFor
//
// Noted This function must be called with
// Lock ALREADY ACQUIRED
func (rf *Raft) stepdownIfNecessary(higherTerm int) {
	if higherTerm > rf.currentTerm {
		log.Printf("server %d stepping down\n", rf.me)
		// if it's a leader and currentTerm changes, step down
		switch rf.state {
		case Leader:
			log.Printf("server %d is a leader, sending to stepdownchan\n", rf.me)
			rf.stepDownChan <- struct{}{}
			log.Printf("server %d is a leader, sending to stepdownchan successfully\n", rf.me)
		case Candidate:
			log.Printf("server %d is a candidate, sending to stepdownchan\n", rf.me)
			rf.stepDownChan <- struct{}{}
			log.Printf("server %d is a candidate, sending to stepdownchan successfully\n", rf.me)
		}

		rf.currentTerm = higherTerm
		rf.state = Follower
		rf.votedFor = -1
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.Lock()
	defer func() {
		rf.Unlock()
		log.Printf("server %d AppendEntriesRPC exiting\n", rf.me)
	}()

	log.Printf("server %d AppendEntriesRPC executing\n", rf.me)

	// Reply false if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Sucess = false
		log.Printf("server %d AppendEntriesRPC agrs Term is less than currentTerm\n", rf.me)
		return
	}
	// valid leader, step down if necessary
	rf.stepdownIfNecessary(args.Term)

	// valid leader, reset election timeout
	select {
	case rf.appendEntriesChan <- struct{}{}:
	default:
	}

	// Reply false if log doesn’t contain an entry at prevLogIndex
	// whose term matches prevLogTerm
	if args.PrevLogIndex >= 0 && len(rf.logs) > 0 {
		if len(rf.logs)-1 < args.PrevLogIndex || rf.logs[args.PrevLogIndex].Term != args.PrevLogTerm {
			log.Printf("server %d AppendEntriesRPC\n", rf.me)
			reply.Term = rf.currentTerm
			reply.Sucess = false
			return
		}
	}

	// If an existing entry conflicts with a new one (same index
	// 	but different terms), delete the existing entry and all that
	// 	follow it
	insertIndex := args.PrevLogIndex + 1
	entriesIndex := 0
	log.Printf("server %d AppendEntriesRPC entries %#v\n", rf.me, args.Entries)

	for ; insertIndex < len(rf.logs) && entriesIndex < len(args.Entries); insertIndex++ {
		log.Printf("server %d AppendEntriesRPC insertIndex %d entriesIndex %d\n", rf.me, insertIndex, entriesIndex)
		if rf.logs[insertIndex].Term != args.Entries[entriesIndex].Term {
			log.Printf("server %d AppendEntriesRPC log before deletion %#v\n", rf.me, rf.logs)
			rf.logs = rf.logs[:insertIndex]
			log.Printf("server %d AppendEntriesRPC log after deletion %#v\n", rf.me, rf.logs)
			break
		}
		entriesIndex++
	}

	// Append any new entries not already in the log
	log.Printf("server %d AppendEntriesRPC log before append %#v\n", rf.me, rf.logs)
	rf.logs = append(rf.logs, args.Entries[entriesIndex:]...)
	log.Printf("server %d AppendEntriesRPC log after append %#v\n", rf.me, rf.logs)

	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = Min(args.LeaderCommit, len(rf.logs)-1)
		log.Printf("server %d AppendEntriesRPC new commitIndex is %d\n", rf.me, rf.commitIndex)
		rf.readyToCommitChan <- struct{}{}
	}

	reply.Sucess = true
	reply.Term = args.Term
}

// example RequestVote RPC handler.
// Invoked by candiates to other nodes to gather votes
// the receiver rf is the other nodes, not the candidate's node
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.Lock()
	defer func() {
		rf.Unlock()
		log.Printf("server %d RequestVoteRPC exiting\n", rf.me)
	}()
	log.Printf("server %d RequestVoteRPC executing\n", rf.me)

	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}
	rf.stepdownIfNecessary(args.Term)

	if args.Term == rf.currentTerm {
		// If votedFor is null or candidateId
		if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
			var (
				lastLogIndex int
				lastLogTerm  int
			)
			if i := len(rf.logs); i > 0 {
				lastLogIndex = i - 1
				lastLogTerm = rf.logs[lastLogIndex].Term
			}

			// candidate’s log is at least as up-to-date as receiver’s log, grant vote
			if args.LastLogTerm > lastLogTerm || args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex {
				// update currentTerm because candidate's term is larger than currentTerm

				reply.Term = args.Term
				reply.VoteGranted = true
				rf.votedFor = args.CandidateId
				select {
				case rf.grantVoteChan <- struct{}{}:
				default:
				}
				log.Printf("server %d granted vote to server %d\n", rf.me, args.CandidateId)
			}
		}
	}
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

func (rf *Raft) isMajority(count int) bool {
	return count >= len(rf.peers)/2+1
}

func (rf *Raft) sendAppendEntriesToAllServers(sendEmpty bool) {
	rf.Lock()
	log.Printf("server %d sending AppendEntries to all servers\n", rf.me)
	defer func() {
		rf.Unlock()
		log.Printf("server %d exiting sendAppendEntriesToAllServers\n", rf.me)
	}()

	ch := make(chan AppendEntriesReply, len(rf.peers))
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		go func(i int) {
			rf.Lock()
			defer rf.Unlock()
			// leader has already step down
			if rf.state != Leader {
				return
			}

			var (
				prevLogTerm, prevLogIndex int
				entries                   []LogEntry
			)
			lastLogIndex := len(rf.logs) - 1
			prevLogIndex = rf.nextIndex[i] - 1
			if prevLogIndex < 0 {
				prevLogIndex = 0
			}
			prevLogTerm = rf.logs[prevLogIndex].Term
			log.Printf("server %d has server %d prevLogIndex %d prevLogTerm %d\n", rf.me, i, prevLogIndex, prevLogTerm)
			// If last log index ≥ nextIndex for a follower: send
			// AppendEntries RPC with log entries starting at nextIndex
			if sendEmpty {
				entries = []LogEntry{}
			} else if lastLogIndex >= rf.nextIndex[i] {
				entries = make([]LogEntry, len(rf.logs[rf.nextIndex[i]:]))
				copy(entries, rf.logs[rf.nextIndex[i]:])
			}
			args := AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderID:     rf.me,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      entries,
				LeaderCommit: rf.commitIndex,
			}

			reply := AppendEntriesReply{}
			log.Printf("server %d sending AppendEntries to server %d\n", rf.me, i)

			// sendAppendEntries is going to take a while, release lock
			rf.Unlock()
			ok := rf.sendAppendEntries(i, &args, &reply)
			rf.Lock()
			// leader has already step down
			if rf.state != Leader {
				return
			}

			if !ok {
				log.Printf("server %d AppendEntries reply from server %d arrives late\n", rf.me, i)
			} else {
				log.Printf("server %d received AppendEntries reply from server %d\n", rf.me, i)
				// If successful: update nextIndex and matchIndex for follower
				if reply.Sucess {
					rf.nextIndex[i] += len(entries)
					rf.matchIndex[i] = rf.nextIndex[i] - 1
					log.Printf("server %d has server %d nextIndex %d matchIndex %d\n", rf.me, i, rf.nextIndex[i], rf.matchIndex[i])

				} else {
					if reply.Term == rf.currentTerm {
						// If AppendEntries fails because of log inconsistency
						// decrement nextIndex and retry
						log.Printf("server %d AppendEntries from server %d fails because of log inconsistency next index is %d\n", rf.me, i, rf.nextIndex[i]-1)
						rf.nextIndex[i]--
					}
				}
			}

			rf.Unlock()
			ch <- reply
			rf.Lock()
		}(i)
	}

	count := 1
	failCount := 1
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		rf.Unlock()
		log.Printf("server %d AppendEntries waiting for reply\n", rf.me)
		reply := <-ch
		log.Printf("server %d AppendEntries received reply\n", rf.me)
		rf.Lock()
		// If RPC request or response contains term T > currentTerm:
		// set currentTerm = T, convert to follower
		if reply.Term > rf.currentTerm {
			rf.stepdownIfNecessary(reply.Term)
			return
		}
		// leader has already step down
		if rf.state != Leader {
			return
		}

		// increase count for AppendEntries received
		count++
		if !reply.Sucess {
			// already checked if reply.Term > rf.currentTerm
			// so Success false should be log inconsistency
			failCount++
		}
		if rf.isMajority(count) {
			// receive majority of AppendEntries
			if rf.isMajority(failCount) {
				log.Printf("server %d received majority false AppendEntries %d, can't commit\n", rf.me, failCount)
				return
			}

			log.Printf("server %d received majority AppendEntries %d\n", rf.me, count)
			// received majority of AppendEntries, check if it can be commited
			oldCommitIndex := rf.commitIndex
			for N := rf.commitIndex + 1; N < len(rf.logs); N++ {
				if rf.logs[N].Term == rf.currentTerm {
					// If there exists an N such that N > commitIndex, a majority
					// of matchIndex[i] ≥ N, and log[N].term == currentTerm:
					// set commitIndex = N
					matchIndexCount := 1
					for i := range rf.peers {
						if i == rf.me {
							continue
						}
						if rf.matchIndex[i] >= N {
							matchIndexCount++
						}
					}

					if rf.isMajority(matchIndexCount) {
						log.Printf("server %d received majority matchIndexCount %d commiting the message", rf.me, matchIndexCount)
						rf.commitIndex = N
					}
				}
			}

			if oldCommitIndex != rf.commitIndex {
				rf.readyToCommitChan <- struct{}{}
			}
		}
	}
}

// beginElection return a channel which will be sent
// a value if it wins the election
func (rf *Raft) beginElection() <-chan struct{} {
	elected := make(chan struct{}, 1)

	go func() {
		rf.Lock()
		// Candidate has already step down
		if rf.state != Candidate {
			return
		}
		log.Printf("server %d begins election\n", rf.me)
		defer func() {
			rf.Unlock()
			log.Printf("server %d end of election\n", rf.me)
		}()

		var (
			lastLogIndex int
			lastLogTerm  int
		)

		if i := len(rf.logs); i > 0 {
			lastLogIndex = i - 1
			lastLogTerm = rf.logs[i-1].Term
		}

		args := RequestVoteArgs{
			Term:         rf.currentTerm,
			CandidateId:  rf.me,
			LastLogIndex: lastLogIndex,
			LastLogTerm:  lastLogTerm,
		}

		ch := make(chan RequestVoteReply, len(rf.peers))
		for i := range rf.peers {
			if i == rf.me {
				continue
			}
			go func(i int) {
				reply := RequestVoteReply{}
				log.Printf("server %d sending vote request to server %d\n", rf.me, i)
				ok := rf.sendRequestVote(i, &args, &reply)
				if !ok {
					log.Printf("server %d VoteRequest reply from server %d arrives late\n", rf.me, i)
				} else {
					log.Printf("server %d received VoteRequest voteGranted %t from server %d\n", rf.me, reply.VoteGranted, i)
				}

				ch <- reply
			}(i)
		}

		count := 1
		for i := range rf.peers {
			if i == rf.me {
				continue
			}
			// release lock because this is going to take a while
			rf.Unlock()
			log.Printf("server %d VoteRequest waiting for reply\n", rf.me)
			reply := <-ch
			log.Printf("server %d VoteRequest receive reply\n", rf.me)
			rf.Lock()
			// If RPC request or response contains term T > currentTerm:
			// set currentTerm = T, convert to follower
			if reply.Term > rf.currentTerm {
				rf.stepdownIfNecessary(reply.Term)
				return
			}
			// Candidate has already step down
			if rf.state != Candidate {
				return
			}
			if reply.VoteGranted {
				count++
			}
			if rf.isMajority(count) {
				log.Printf("server %d received majority votes %d\n", rf.me, count)
				elected <- struct{}{}
				return
			}
		}
	}()

	return elected
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := false

	// Your code here (2B).
	rf.Lock()
	defer rf.Unlock()

	if rf.state == Leader {
		index = len(rf.logs)
		term = rf.currentTerm
		isLeader = true

		rf.logs = append(rf.logs,
			LogEntry{
				Command: command,
				Term:    term,
			})
		log.Printf("server %d appended logs %#v in term %d\n", rf.me, command, term)

	}

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) generateRandomElectionTimeout() time.Duration {
	randRange := rand.Intn(electionTimeoutMax - electionTimeoutMin)
	randElectionTimeout := time.Duration(electionTimeoutMin+randRange) * time.Millisecond
	return randElectionTimeout
}

func (rf *Raft) commit() {
	for range rf.readyToCommitChan {
		log.Printf("server %d ready to commit\n", rf.me)
		rf.Lock()
		lastApplied := rf.lastApplied
		commitIndex := rf.commitIndex
		rf.Unlock()

		for i := lastApplied + 1; i <= commitIndex; i++ {
			rf.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      rf.logs[i].Command,
				CommandIndex: i,
			}
		}

		rf.Lock()
		log.Printf("server %d lastApplied %d\n", rf.me, commitIndex)
		rf.lastApplied = commitIndex
		rf.Unlock()
	}
}

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for !rf.killed() {
		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().
		rf.Lock()

		switch rf.state {
		case Follower:
			// If election timeout elapses without receiving AppendEntries
			// RPC from current leader or granting vote to candidate:
			// convert to candidate
			log.Printf("server %d is a Follower\n", rf.me)
			rf.Unlock()
			select {
			case <-time.After(rf.generateRandomElectionTimeout()):
				// acquire lock will need some time and within this
				// time periods, anything can happen
				rf.Lock()
				log.Printf("server %d is checking channel before switching to Follower\n", rf.me)
				select {
				case <-rf.appendEntriesChan:
					log.Printf("server %d receive appendEntries, resetting timeout\n", rf.me)
				case <-rf.grantVoteChan:
					log.Printf("server %d granted vote, resetting timeout\n", rf.me)
				default:
					log.Printf("server %d converted to candidate\n", rf.me)
					rf.state = Candidate
				}
				rf.Unlock()
			case <-rf.appendEntriesChan:
				log.Printf("server %d receive appendEntries, resetting timeout\n", rf.me)
			case <-rf.grantVoteChan:
				log.Printf("server %d granted vote, resetting timeout\n", rf.me)
			}
			rf.Lock()
		case Candidate:
			// On conversion to candidate, start election
			log.Printf("server %d is a Candidate\n", rf.me)

			rf.currentTerm++
			rf.votedFor = rf.me

			rf.Unlock()
			select {
			case <-rf.beginElection():
				// acquire lock will need some time and within this
				// time periods, anything can happen
				rf.Lock()
				log.Printf("server %d is checking channel before switching to Follower\n", rf.me)
				select {
				case <-rf.appendEntriesChan:
					log.Printf("server %d received valid AppendEntries as candidate, stepping down\n", rf.me)
					rf.state = Follower
				case <-rf.stepDownChan:
					// discover higher term
					log.Printf("server %d is stepping down as a Candidate\n", rf.me)
				default:
					log.Printf("server %d became leader\n", rf.me)
					rf.state = Leader
					// Initializes all nextIndex values to the index just after the last one in its log
					for i := range rf.peers {
						rf.nextIndex[i] = len(rf.logs)
					}
				}
				rf.Unlock()
				// 	Upon election: send initial empty AppendEntries RPCs
				// (heartbeat) to each server
				go rf.sendAppendEntriesToAllServers(true)

			case <-time.After(rf.generateRandomElectionTimeout()):
				log.Printf("server %d election timeout is expired\n", rf.me)
			// If AppendEntries RPC received from new leader: step down
			case <-rf.appendEntriesChan:
				rf.Lock()
				log.Printf("server %d received valid AppendEntries as candidate, stepping down\n", rf.me)
				rf.state = Follower
				rf.Unlock()
			case <-rf.stepDownChan:
				// discover higher term
				rf.Lock()
				log.Printf("server %d is stepping down as a Candidate\n", rf.me)
				rf.state = Follower
				rf.Unlock()
			}
			rf.Lock()

		case Leader:
			log.Printf("server %d is a Leader\n", rf.me)
			rf.Unlock()

			// If step down, immediately go to another state
			select {
			case <-rf.stepDownChan:
				log.Printf("server %d is stepping down as a Leader\n", rf.me)
			case <-time.After(heartbeatInterval):
				go rf.sendAppendEntriesToAllServers(false)
			}

			rf.Lock()
		}

		rf.Unlock()
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	// initialize to 0 even though Go will initialize to 0
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	rf.currentTerm = 0
	rf.votedFor = -1
	rf.logs = []LogEntry{{}}
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	rf.state = Follower
	rf.appendEntriesChan = make(chan struct{}, 1)
	rf.grantVoteChan = make(chan struct{}, 1)
	rf.stepDownChan = make(chan struct{}, 1)
	rf.readyToCommitChan = make(chan struct{})
	rf.applyCh = applyCh
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.commit()
	go rf.ticker()

	return rf
}
