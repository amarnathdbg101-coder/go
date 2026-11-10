package crdt

// Config holds common settings for CRDT operations.
type Config struct {
	// NodeID is the unique identifier of the replica (e.g. "pos_terminal_1", "server_cluster_a").
	NodeID string
}

// Option is a functional option for configuring CRDT instances.
type Option func(*Config)

// WithNodeID specifies a custom node identifier.
func WithNodeID(nodeID string) Option {
	return func(c *Config) {
		c.NodeID = nodeID
	}
}
