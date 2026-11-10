package raft

import (
	"math/rand"
	"sync"
	"time"
)

// Transport abstracts network communication between Raft cluster peers.
type Transport interface {
	SendRequestVote(target string, args *RequestVoteArgs) (*RequestVoteReply, error)
	SendAppendEntries(target string, args *AppendEntriesArgs) (*AppendEntriesReply, error)
}

// Node represents a Raft consensus node.
type Node struct {
	mu        sync.Mutex
	cfg       Config
	transport Transport
	role      Role

	// Persistent state on all servers
	currentTerm uint64
	votedFor    string
	log         []LogEntry // 1-indexed conceptually: log[0] is dummy entry

	// Volatile state on all servers
	commitIndex uint64
	lastApplied uint64

	// Volatile state on leaders
	nextIndex  map[string]uint64
	matchIndex map[string]uint64

	applyCh            chan ApplyMsg
	stopChan           chan struct{}
	electionResetTime  time.Time
	rnd                *rand.Rand
}

// New creates and starts a Raft consensus node.
func New(id string, peers []string, transport Transport, applyCh chan ApplyMsg, opts ...Option) (*Node, error) {
	cfg := DefaultConfig(id, peers)
	for _, opt := range opts {
		opt(&cfg)
	}

	n := &Node{
		cfg:               cfg,
		transport:         transport,
		role:              RoleFollower,
		log:               []LogEntry{{Index: 0, Term: 0}}, // Dummy entry at index 0
		nextIndex:         make(map[string]uint64),
		matchIndex:        make(map[string]uint64),
		applyCh:           applyCh,
		stopChan:          make(chan struct{}),
		electionResetTime: time.Now(),
		rnd:               rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	go n.runElectionLoop()
	return n, nil
}

// GetState returns current term and whether this node believes it is the leader.
func (n *Node) GetState() (uint64, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.currentTerm, n.role == RoleLeader
}

// Propose submits a command to the cluster. Returns log index and term if accepted as leader.
func (n *Node) Propose(cmd []byte) (uint64, uint64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	select {
	case <-n.stopChan:
		return 0, 0, ErrNodeStopped
	default:
	}

	if n.role != RoleLeader {
		return 0, 0, ErrNotLeader
	}

	idx := uint64(len(n.log))
	entry := LogEntry{
		Index:   idx,
		Term:    n.currentTerm,
		Command: cmd,
	}
	n.log = append(n.log, entry)
	n.matchIndex[n.cfg.ID] = idx

	go n.broadcastAppendEntries()
	return idx, n.currentTerm, nil
}

// HandleRequestVote processes an incoming RequestVote RPC.
func (n *Node) HandleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term > n.currentTerm {
		n.becomeFollower(args.Term)
	}

	reply.Term = n.currentTerm
	reply.VoteGranted = false

	if args.Term < n.currentTerm {
		return
	}

	canVote := n.votedFor == "" || n.votedFor == args.CandidateID
	lastLog := n.log[len(n.log)-1]

	// Check if candidate's log is at least as up-to-date as receiver's
	isUpToDate := args.LastLogTerm > lastLog.Term ||
		(args.LastLogTerm == lastLog.Term && args.LastLogIndex >= lastLog.Index)

	if canVote && isUpToDate {
		reply.VoteGranted = true
		n.votedFor = args.CandidateID
		n.electionResetTime = time.Now()
	}
}

// HandleAppendEntries processes an incoming AppendEntries (heartbeat/replication) RPC.
func (n *Node) HandleAppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term > n.currentTerm {
		n.becomeFollower(args.Term)
	}

	reply.Term = n.currentTerm
	reply.Success = false

	if args.Term < n.currentTerm {
		return
	}

	n.electionResetTime = time.Now()
	if n.role != RoleFollower {
		n.becomeFollower(args.Term)
	}

	// Verify log contains an entry at PrevLogIndex matching PrevLogTerm
	if args.PrevLogIndex >= uint64(len(n.log)) || n.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		return
	}

	// Append any new entries not already in the log
	for i, entry := range args.Entries {
		insertIdx := args.PrevLogIndex + 1 + uint64(i)
		if insertIdx < uint64(len(n.log)) {
			if n.log[insertIdx].Term != entry.Term {
				n.log = n.log[:insertIdx] // truncate mismatch
				n.log = append(n.log, entry)
			}
		} else {
			n.log = append(n.log, entry)
		}
	}

	reply.Success = true
	reply.MatchIndex = uint64(len(n.log) - 1)

	// Update commitIndex if leader has committed
	if args.LeaderCommit > n.commitIndex {
		lastIdx := uint64(len(n.log) - 1)
		if args.LeaderCommit < lastIdx {
			n.commitIndex = args.LeaderCommit
		} else {
			n.commitIndex = lastIdx
		}
		n.applyLogs()
	}
}

func (n *Node) becomeFollower(term uint64) {
	n.role = RoleFollower
	n.currentTerm = term
	n.votedFor = ""
	n.electionResetTime = time.Now()
}

func (n *Node) runElectionLoop() {
	for {
		timeout := n.cfg.ElectionTimeout + time.Duration(n.rnd.Intn(150))*time.Millisecond
		time.Sleep(50 * time.Millisecond)

		n.mu.Lock()
		select {
		case <-n.stopChan:
			n.mu.Unlock()
			return
		default:
		}

		if n.role != RoleLeader && time.Since(n.electionResetTime) >= timeout {
			n.startElection()
		}
		n.mu.Unlock()
	}
}

func (n *Node) startElection() {
	n.role = RoleCandidate
	n.currentTerm++
	n.votedFor = n.cfg.ID
	n.electionResetTime = time.Now()

	lastLog := n.log[len(n.log)-1]
	args := RequestVoteArgs{
		Term:         n.currentTerm,
		CandidateID:  n.cfg.ID,
		LastLogIndex: lastLog.Index,
		LastLogTerm:  lastLog.Term,
	}

	votes := 1 // Vote for self
	totalNodes := len(n.cfg.Peers) + 1
	needed := totalNodes/2 + 1

	for _, peer := range n.cfg.Peers {
		go func(target string) {
			if n.transport == nil {
				return
			}
			reply, err := n.transport.SendRequestVote(target, &args)
			if err != nil || reply == nil {
				return
			}

			n.mu.Lock()
			defer n.mu.Unlock()

			if reply.Term > n.currentTerm {
				n.becomeFollower(reply.Term)
				return
			}

			if n.role == RoleCandidate && reply.VoteGranted && reply.Term == n.currentTerm {
				votes++
				if votes >= needed {
					n.becomeLeader()
				}
			}
		}(peer)
	}
}

func (n *Node) becomeLeader() {
	n.role = RoleLeader
	lastIdx := uint64(len(n.log) - 1)
	for _, peer := range n.cfg.Peers {
		n.nextIndex[peer] = lastIdx + 1
		n.matchIndex[peer] = 0
	}
	n.matchIndex[n.cfg.ID] = lastIdx

	go n.runHeartbeatLoop()
	go n.broadcastAppendEntries()
}

func (n *Node) runHeartbeatLoop() {
	ticker := time.NewTicker(n.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.mu.Lock()
			if n.role != RoleLeader {
				n.mu.Unlock()
				return
			}
			n.mu.Unlock()
			n.broadcastAppendEntries()
		}
	}
}

func (n *Node) broadcastAppendEntries() {
	n.mu.Lock()
	if n.role != RoleLeader {
		n.mu.Unlock()
		return
	}

	term := n.currentTerm
	leaderID := n.cfg.ID
	leaderCommit := n.commitIndex

	for _, peer := range n.cfg.Peers {
		nextIdx := n.nextIndex[peer]
		if nextIdx == 0 {
			nextIdx = 1
		}
		prevIdx := nextIdx - 1
		prevTerm := n.log[prevIdx].Term

		var entries []LogEntry
		if uint64(len(n.log)) > nextIdx {
			entries = n.log[nextIdx:]
		}

		args := AppendEntriesArgs{
			Term:         term,
			LeaderID:     leaderID,
			PrevLogIndex: prevIdx,
			PrevLogTerm:  prevTerm,
			Entries:      entries,
			LeaderCommit: leaderCommit,
		}

		go func(target string, a AppendEntriesArgs) {
			if n.transport == nil {
				return
			}
			reply, err := n.transport.SendAppendEntries(target, &a)
			if err != nil || reply == nil {
				return
			}

			n.mu.Lock()
			defer n.mu.Unlock()

			if reply.Term > n.currentTerm {
				n.becomeFollower(reply.Term)
				return
			}

			if n.role == RoleLeader && reply.Term == n.currentTerm {
				if reply.Success {
					n.nextIndex[target] = reply.MatchIndex + 1
					n.matchIndex[target] = reply.MatchIndex
					n.checkAndUpdateCommitIndex()
				} else if n.nextIndex[target] > 1 {
					n.nextIndex[target]-- // Step back log on mismatch
				}
			}
		}(peer, args)
	}
	n.mu.Unlock()
}

func (n *Node) checkAndUpdateCommitIndex() {
	totalNodes := len(n.cfg.Peers) + 1
	needed := totalNodes/2 + 1

	for idx := uint64(len(n.log) - 1); idx > n.commitIndex; idx-- {
		if n.log[idx].Term != n.currentTerm {
			continue // Leaders only commit entries from current term
		}

		count := 1 // Leader itself
		for _, peer := range n.cfg.Peers {
			if n.matchIndex[peer] >= idx {
				count++
			}
		}

		if count >= needed {
			n.commitIndex = idx
			n.applyLogs()
			break
		}
	}
}

func (n *Node) applyLogs() {
	for n.commitIndex > n.lastApplied {
		n.lastApplied++
		entry := n.log[n.lastApplied]
		if n.applyCh != nil && len(entry.Command) > 0 {
			n.applyCh <- ApplyMsg{
				CommandValid: true,
				Command:      entry.Command,
				CommandIndex: entry.Index,
			}
		}
	}
}

// Stop gracefully shuts down the Raft node.
func (n *Node) Stop() {
	n.mu.Lock()
	defer n.mu.Unlock()
	select {
	case <-n.stopChan:
		return
	default:
		close(n.stopChan)
	}
}
