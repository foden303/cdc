package utils

import (
	"strings"
	"unsafe"

	"github.com/spaolacci/murmur3"
)

// GeneratePartition calculates a partition ID for a given key string.
// It uses MurmurHash3 (32-bit) for highly uniform distribution.
// If the key is empty, it returns 0.
func GeneratePartition(key string, partitionCount int) int {
	if key == "" || partitionCount <= 1 {
		return 0
	}
	// MurmurHash3 is statistically stronger than CRC32 for small bucket partitioning
	checksum := murmur3.Sum32(unsafe.Slice(unsafe.StringData(key), len(key)))

	return int(checksum % uint32(partitionCount))
}

// CombineKeys joins multiple strings with a separator to create a composite partition key.
func CombineKeys(keys ...string) string {
	if len(keys) == 0 {
		return ""
	}

	return strings.Join(keys, ":")
}
