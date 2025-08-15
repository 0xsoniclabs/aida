package state

import "github.com/0xsoniclabs/carmen/go/carmen"

//go:generate mockgen -source carmen_test_helper.go -destination carmen_mock.go -package state

type proxyDatabase interface {
	carmen.Database
}

type proxyTransactionContext interface {
	carmen.TransactionContext
}

type proxyHistoricBlockContext interface {
	carmen.HistoricBlockContext
}

type proxyQueryContext interface {
	carmen.QueryContext
}

type proxyMemoryFootprint interface {
	carmen.MemoryFootprint
}
