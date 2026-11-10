package raft

import "errors"

var (
	// ErrNotLeader is returned when a client command is proposed to a non-leader node.
	ErrNotLeader = errors.New("raft: node is not the current cluster leader")

	// ErrNodeStopped is returned when proposing an action to a stopped node.
	ErrNodeStopped = errors.New("raft: node has been stopped")

	// ErrStaleTerm is returned when an incoming RPC has an older term than the local node.
	ErrStaleTerm = errors.New("raft: incoming term is stale")

	// ErrLogMismatch is returned when follower's log does not match leader's PrevLogIndex/PrevLogTerm.
	ErrLogMismatch = errors.New("raft: log consistency check failed")
)
