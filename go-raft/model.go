package raft

// Role defines the current state of a Raft node.
type Role int

const (
	// RoleFollower is the initial state of all nodes.
	RoleFollower Role = iota
	// RoleCandidate is assumed when election timer expires without heartbeat.
	RoleCandidate
	// RoleLeader manages log replication and sends heartbeats.
	RoleLeader
)

func (r Role) String() string {
	switch r {
	case RoleFollower:
		return "Follower"
	case RoleCandidate:
		return "Candidate"
	case RoleLeader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// LogEntry represents an indexed command in the replicated log.
type LogEntry struct {
	Index   uint64
	Term    uint64
	Command []byte
}

// ApplyMsg represents a committed entry delivered to application state machine.
type ApplyMsg struct {
	CommandValid bool
	Command      []byte
	CommandIndex uint64
}

// RequestVoteArgs payload.
type RequestVoteArgs struct {
	Term         uint64
	CandidateID  string
	LastLogIndex uint64
	LastLogTerm  uint64
}

// RequestVoteReply payload.
type RequestVoteReply struct {
	Term        uint64
	VoteGranted bool
}

// AppendEntriesArgs payload (heartbeat & replication).
type AppendEntriesArgs struct {
	Term         uint64
	LeaderID     string
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []LogEntry
	LeaderCommit uint64
}

// AppendEntriesReply payload.
type AppendEntriesReply struct {
	Term       uint64
	Success    bool
	MatchIndex uint64
}
