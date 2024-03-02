// Copyright 2019+ Klaus Post. All rights reserved.
// License information can be found in the LICENSE file.
// Based on work by Yann Collet, released under BSD License.

package zstd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
)

// bitReader reads a bitstream in reverse.
// The last set bit indicates the start of the stream and is used
// for aligning the input.
type bitReader struct {
	in       []byte
<<<<<<< HEAD
<<<<<<< HEAD
=======
	off      uint   // next byte to read is at in[off - 1]
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	value    uint64 // Maybe use [16]byte, but shifting is awkward.
	bitsRead uint8
}

// init initializes and resets the bit reader.
func (b *bitReader) init(in []byte) error {
	if len(in) < 1 {
		return errors.New("corrupt stream: too short")
	}
	b.in = in
<<<<<<< HEAD
<<<<<<< HEAD
=======
	b.off = uint(len(in))
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	// The highest bit of the last byte indicates where to start
	v := in[len(in)-1]
	if v == 0 {
		return errors.New("corrupt stream, did not find end of stream")
	}
	b.bitsRead = 64
	b.value = 0
	if len(in) >= 8 {
		b.fillFastStart()
	} else {
		b.fill()
		b.fill()
	}
	b.bitsRead += 8 - uint8(highBits(uint32(v)))
	return nil
}

// getBits will return n bits. n can be 0.
func (b *bitReader) getBits(n uint8) int {
	if n == 0 /*|| b.bitsRead >= 64 */ {
		return 0
	}
	return int(b.get32BitsFast(n))
}

// get32BitsFast requires that at least one bit is requested every time.
// There are no checks if the buffer is filled.
func (b *bitReader) get32BitsFast(n uint8) uint32 {
	const regMask = 64 - 1
	v := uint32((b.value << (b.bitsRead & regMask)) >> ((regMask + 1 - n) & regMask))
	b.bitsRead += n
	return v
}

// fillFast() will make sure at least 32 bits are available.
// There must be at least 4 bytes available.
func (b *bitReader) fillFast() {
	if b.bitsRead < 32 {
		return
	}
<<<<<<< HEAD
<<<<<<< HEAD
	v := b.in[len(b.in)-4:]
	b.in = b.in[:len(b.in)-4]
	low := (uint32(v[0])) | (uint32(v[1]) << 8) | (uint32(v[2]) << 16) | (uint32(v[3]) << 24)
	b.value = (b.value << 32) | uint64(low)
	b.bitsRead -= 32
=======
	// 2 bounds checks.
	v := b.in[b.off-4:]
	v = v[:4]
	low := (uint32(v[0])) | (uint32(v[1]) << 8) | (uint32(v[2]) << 16) | (uint32(v[3]) << 24)
	b.value = (b.value << 32) | uint64(low)
	b.bitsRead -= 32
	b.off -= 4
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	v := b.in[len(b.in)-4:]
	b.in = b.in[:len(b.in)-4]
	low := (uint32(v[0])) | (uint32(v[1]) << 8) | (uint32(v[2]) << 16) | (uint32(v[3]) << 24)
	b.value = (b.value << 32) | uint64(low)
	b.bitsRead -= 32
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}

// fillFastStart() assumes the bitreader is empty and there is at least 8 bytes to read.
func (b *bitReader) fillFastStart() {
<<<<<<< HEAD
<<<<<<< HEAD
	v := b.in[len(b.in)-8:]
	b.in = b.in[:len(b.in)-8]
	b.value = binary.LittleEndian.Uint64(v)
	b.bitsRead = 0
=======
	// Do single re-slice to avoid bounds checks.
	b.value = binary.LittleEndian.Uint64(b.in[b.off-8:])
	b.bitsRead = 0
	b.off -= 8
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	v := b.in[len(b.in)-8:]
	b.in = b.in[:len(b.in)-8]
	b.value = binary.LittleEndian.Uint64(v)
	b.bitsRead = 0
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}

// fill() will make sure at least 32 bits are available.
func (b *bitReader) fill() {
	if b.bitsRead < 32 {
		return
	}
<<<<<<< HEAD
<<<<<<< HEAD
	if len(b.in) >= 4 {
		v := b.in[len(b.in)-4:]
		b.in = b.in[:len(b.in)-4]
		low := (uint32(v[0])) | (uint32(v[1]) << 8) | (uint32(v[2]) << 16) | (uint32(v[3]) << 24)
		b.value = (b.value << 32) | uint64(low)
		b.bitsRead -= 32
		return
	}

	b.bitsRead -= uint8(8 * len(b.in))
	for len(b.in) > 0 {
		b.value = (b.value << 8) | uint64(b.in[len(b.in)-1])
		b.in = b.in[:len(b.in)-1]
=======
	if b.off >= 4 {
		v := b.in[b.off-4:]
		v = v[:4]
=======
	if len(b.in) >= 4 {
		v := b.in[len(b.in)-4:]
		b.in = b.in[:len(b.in)-4]
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		low := (uint32(v[0])) | (uint32(v[1]) << 8) | (uint32(v[2]) << 16) | (uint32(v[3]) << 24)
		b.value = (b.value << 32) | uint64(low)
		b.bitsRead -= 32
		return
	}
<<<<<<< HEAD
	for b.off > 0 {
		b.value = (b.value << 8) | uint64(b.in[b.off-1])
		b.bitsRead -= 8
		b.off--
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======

	b.bitsRead -= uint8(8 * len(b.in))
	for len(b.in) > 0 {
		b.value = (b.value << 8) | uint64(b.in[len(b.in)-1])
		b.in = b.in[:len(b.in)-1]
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	}
}

// finished returns true if all bits have been read from the bit stream.
func (b *bitReader) finished() bool {
<<<<<<< HEAD
<<<<<<< HEAD
	return len(b.in) == 0 && b.bitsRead >= 64
=======
	return b.off == 0 && b.bitsRead >= 64
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	return len(b.in) == 0 && b.bitsRead >= 64
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}

// overread returns true if more bits have been requested than is on the stream.
func (b *bitReader) overread() bool {
	return b.bitsRead > 64
}

// remain returns the number of bits remaining.
func (b *bitReader) remain() uint {
<<<<<<< HEAD
<<<<<<< HEAD
	return 8*uint(len(b.in)) + 64 - uint(b.bitsRead)
=======
	return b.off*8 + 64 - uint(b.bitsRead)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	return 8*uint(len(b.in)) + 64 - uint(b.bitsRead)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}

// close the bitstream and returns an error if out-of-buffer reads occurred.
func (b *bitReader) close() error {
	// Release reference.
	b.in = nil
	if !b.finished() {
		return fmt.Errorf("%d extra bits on block, should be 0", b.remain())
	}
	if b.bitsRead > 64 {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func highBits(val uint32) (n uint32) {
	return uint32(bits.Len32(val) - 1)
}
