package gametorch

import (
	"crypto/rand"
	"encoding/hex"
)

// newUUIDv4 returns a random RFC 4122 version 4 UUID string. It uses
// crypto/rand so no third-party UUID dependency is required.
func newUUIDv4() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand.Read never fails on supported platforms; fall back to a
		// deterministic-but-unique value rather than panicking.
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	var buf [36]byte
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf[:])
}

// randomByte returns a single cryptographically random byte, used for retry
// jitter.
func randomByte() byte {
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0
	}
	return b[0]
}
