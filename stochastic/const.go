package stochastic

// Argument classifier for StateDB arguments (account addresses, storage keys,  and storage values).
// The  classifier is based on a combination of counting and queuing statistics to classify arguments
// into the following kinds:
const (
	NoArgID     = iota // no argument
	ZeroArgID          // zero argument
	PrevArgID          // previously seen argument (last access)
	RecentArgID        // recent argument value (found in the counting queue)
	RandArgID          // random access (pick randomly from argument set)
	NewArgID           // new argument (not seen before and increases the argument set)

	NumArgKinds // number of argument kinds
)

// QueueLen sets the length of counting queue (must be greater than one).
const QueueLen = 32

// NumECDFPoints sets the number of points in the empirical cumulative distribution function.
const NumECDFPoints = 300
