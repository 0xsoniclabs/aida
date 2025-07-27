package state

import "github.com/0xsoniclabs/carmen/go/carmen"

//go:generate mockgen -source carmen_test_helper.go -destination carmen_mock.go -package state

// nolint: unused
type proxyMemoryFootprint interface {
	carmen.MemoryFootprint
}
