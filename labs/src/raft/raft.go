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
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"../labrpc"
)

// import "bytes"
// import "../labgob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

const (
	LEADER    = 0
	CANDIDATE = 1
	FOLLOWER  = 2

	HEARTBEAT            = 200 * time.Millisecond
	ELECTION_TIMEOUT_MIN = 300
	ELECTION_TIMEOUT_MAX = 500
)

type Log struct {
	Term    int
	Command interface{}
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	state int

	// persistent state on all servers
	currentTerm int
	votedFor    int
	log         []Log

	// volatile state on all servers
	commitIndex int
	lastApplied int

	// volatile state on leaders
	nextIndex  []int
	matchIndex []int

	// additional fields that are not mentioned in paper
	electionTimeout time.Duration
	lastHeard       time.Time // last time heard from the leader
	voteCount       int       // how many votes does a candidate get from an election
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	rf.mu.Lock()
	term = rf.currentTerm
	isleader = rf.state == LEADER
	rf.mu.Unlock()
	return term, isleader
}

func (rf *Raft) resetTimer() {
	rf.mu.Lock()
	interval := rand.Intn(ELECTION_TIMEOUT_MAX-ELECTION_TIMEOUT_MIN) + ELECTION_TIMEOUT_MIN
	rf.electionTimeout = time.Duration(interval) * time.Millisecond
	rf.lastHeard = time.Now()
	rf.mu.Unlock()
}

func (rf *Raft) startHeartBeat() {
	for {
		rf.mu.Lock()
		if rf.state != LEADER {
			rf.mu.Unlock()
			return
		}
		term := rf.currentTerm
		prevLogIndex := len(rf.log) - 1
		prevLogTerm := rf.log[prevLogIndex].Term
		leaderCommit := rf.commitIndex
		rf.mu.Unlock()

		peerNum := len(rf.peers)
		for i := 0; i < peerNum; i++ {
			if i == rf.me {
				continue
			}
			args := AppendEntriesArgs{term, rf.me, prevLogIndex, prevLogTerm, []Log{}, leaderCommit}
			reply := AppendEntriesReply{}
			go rf.sendAppendEntries(i, &args, &reply)
		}
		time.Sleep(HEARTBEAT)
	}
}

func (rf *Raft) leader() {
	// comes to power, initialize fields
	rf.mu.Lock()
	rf.state = LEADER
	rf.votedFor = -1
	peerNum := len(rf.peers)
	for i := 0; i < peerNum; i++ {
		rf.nextIndex[i] = len(rf.log) // leader last log index + 1
		rf.matchIndex[i] = 0
	}
	rf.mu.Unlock()

	go rf.startHeartBeat()

	// TODO: leader work
	for {
		rf.mu.Lock()
		if rf.state == FOLLOWER {
			rf.mu.Unlock()
			go rf.follower()
			return
		}
		rf.mu.Unlock()
	}
}

func (rf *Raft) candidate() {
	rf.mu.Lock()
	rf.currentTerm++
	rf.state = CANDIDATE
	rf.votedFor = rf.me
	rf.voteCount = 1 // vote for itself

	peerNum := len(rf.peers)
	term := rf.currentTerm
	me := rf.me
	lastLogIndex := len(rf.log) - 1
	lastLogTerm := rf.log[lastLogIndex].Term
	rf.mu.Unlock()

	rf.resetTimer()
	go rf.startElectionTimer()

	// send RequestVote to all peers
	for i := 0; i < peerNum; i++ {
		if i == me {
			continue
		}
		args := RequestVoteArgs{term, me, lastLogIndex, lastLogTerm}
		reply := RequestVoteReply{}
		go rf.sendRequestVote(i, &args, &reply)
	}

	// a candidate continues its state until one of three things happens
	// a conditional variable should be used here, but event a) b) and c)
	// are triggered by different goroutines, which increases complexity
	// therefore busy waiting is used here
	for {
		rf.mu.Lock()
		if rf.voteCount > peerNum/2 {
			// a) the candidate wins and becomes leader
			rf.mu.Unlock()
			go rf.leader()
			break
		}
		if rf.state == FOLLOWER {
			// b) another server establishes itself as leader
			rf.mu.Unlock()
			go rf.follower()
			break
		}
		if rf.currentTerm > term {
			// c) a certain peer has already started a new election
			// at this moment, this peer is either running follower() or candidate()
			rf.mu.Unlock()
			break
		}
		rf.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

func (rf *Raft) follower() {
	go rf.startElectionTimer()
	// TODO: follower
}

// election timeout goroutine periodically checks
// whether the time since the last time heard from the leader is greater than the timeout period.
// If so, start a new election and return
// each time a server becomes a follower or starts an election, start this timer goroutine
func (rf *Raft) startElectionTimer() {
	for {
		rf.mu.Lock()
		electionTimeout := rf.electionTimeout
		lastHeard := rf.lastHeard
		rf.mu.Unlock()
		now := time.Now()
		if now.After(lastHeard.Add(electionTimeout)) {
			go rf.candidate()
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
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

//
// restore previously persisted state.
//
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

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		rf.mu.Unlock()
		return
	}
	// follow the second rule in "Rules for Servers" in figure 2 before handling an incoming RPC
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = FOLLOWER
		rf.votedFor = -1
	}

	reply.Term = rf.currentTerm
	reply.VoteGranted = true
	// deny vote if already voted
	if rf.votedFor != -1 {
		reply.VoteGranted = false
		rf.mu.Unlock()
		return
	}
	// deny vote if consistency check fails (candidate is less up-to-date)
	lastLog := rf.log[len(rf.log)-1]
	if args.LastLogTerm < lastLog.Term || (args.LastLogTerm == lastLog.Term && args.LastLogIndex < len(rf.log)-1) {
		reply.VoteGranted = false
		rf.mu.Unlock()
		return
	}
	// now this peer must vote for the candidate
	rf.votedFor = args.CandidateID
	rf.mu.Unlock()

	rf.resetTimer()
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Log
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		rf.mu.Unlock()
		return
	}
	// follow the second rule in "Rules for Servers" in figure 2 before handling an incoming RPC
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = FOLLOWER
		rf.votedFor = -1
	}
	rf.mu.Unlock()
	// now we must have rf.currentTerm == args.Term, which means receiving from leader and reset timer
	rf.resetTimer()

	// consistency check
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if len(rf.log)-1 < args.PrevLogIndex || rf.log[len(rf.log)-1].Term != args.PrevLogTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
		return
	}
	// TODO: log replication
}

//
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
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) {
	if ok := rf.peers[server].Call("Raft.RequestVote", args, reply); !ok {
		return
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if reply.Term > rf.currentTerm {
		rf.currentTerm = reply.Term
		rf.state = FOLLOWER
		return
	}
	if reply.VoteGranted {
		rf.voteCount++
	}
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) {
	for {
		if ok := rf.peers[server].Call("Raft.AppendEntries", args, reply); !ok {
			continue
		}
		rf.mu.Lock()
		defer rf.mu.Unlock()
		if reply.Term > rf.currentTerm {
			rf.currentTerm = reply.Term
			rf.state = FOLLOWER
			return
		}
		if reply.Success {
			rf.matchIndex[server] = args.PrevLogIndex + len(args.Entries)
			rf.nextIndex[server] = rf.matchIndex[server] + 1
			return
		} else {
			return
			// TODO: roll back and send RPC again
		}
	}
}

//
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
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.mu = sync.Mutex{}
	rf.state = FOLLOWER

	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = []Log{}
	// log index counts from 1, index 0 is filled with a fake log with term 0
	rf.log = append(rf.log, Log{0, nil})

	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))

	rf.resetTimer()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start as a follower
	go rf.follower()

	return rf
}
