package state

import "github.com/0xsoniclabs/carmen/go/carmen"

//go:generate mockgen -source carmen_test_helper.go -destination carmen_mock.go -package state

// nolint: unused
type proxyDatabase interface {
	carmen.Database
}

// nolint: unused
type proxyTransactionContext interface {
	carmen.TransactionContext
}

// nolint: unused
type proxyHistoricBlockContext interface {
	carmen.HistoricBlockContext
}

// nolint: unused
type proxyQueryContext interface {
	carmen.QueryContext
}

// nolint: unused
type proxyMemoryFootprint interface {
	carmen.MemoryFootprint
}
