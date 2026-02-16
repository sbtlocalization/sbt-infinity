// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: GPL-3.0-only

// ACM decoder ported from NearInfinity (https://github.com/NearInfinityBrowser/NearInfinity)
// Original Java implementation: Copyright (C) 2001 Jon Olav Hauglid, LGPL 2.1

// SPDX-SnippetBegin
// SPDX-SnippetCopyrightText: Copyright (C) 2001 Jon Olav Hauglid, LGPL 2.1
// SPDX-License-Identifier: LGPL-2.1-only

package sound

import (
	"encoding/binary"
	"fmt"
)

// typedBuf wraps a byte slice with typed element access and a movable offset.
// Mirrors the behavior of NearInfinity's DynamicArray class.
// All multi-byte reads/writes use little-endian byte order.
type typedBuf struct {
	data     []byte
	offset   int // current byte offset
	elemSize int // bytes per logical element: 1, 2, or 4
}

func newTypedBuf(data []byte, offset, elemSize int) typedBuf {
	return typedBuf{data: data, offset: offset, elemSize: elemSize}
}

func allocTypedBuf(count, elemSize int) typedBuf {
	return typedBuf{data: make([]byte, count*elemSize), offset: 0, elemSize: elemSize}
}

func (b typedBuf) clone() typedBuf {
	return typedBuf{data: b.data, offset: b.offset, elemSize: b.elemSize}
}

func (b *typedBuf) addOffset(n int) {
	b.offset += n * b.elemSize
}

func (b *typedBuf) setOffset(byteOfs int) {
	b.offset = byteOfs
}

func (b typedBuf) byteOffset() int {
	return b.offset
}

func (b typedBuf) byteLen() int {
	return len(b.data)
}

func (b typedBuf) getByte(index int) byte {
	return b.data[b.offset+index*b.elemSize]
}

func (b typedBuf) getInt16(index int) int16 {
	pos := b.offset + index*b.elemSize
	return int16(binary.LittleEndian.Uint16(b.data[pos:]))
}

func (b typedBuf) putInt16(index int, v int16) {
	pos := b.offset + index*b.elemSize
	binary.LittleEndian.PutUint16(b.data[pos:], uint16(v))
}

func (b typedBuf) getInt32(index int) int32 {
	pos := b.offset + index*b.elemSize
	return int32(binary.LittleEndian.Uint32(b.data[pos:]))
}

func (b typedBuf) putInt32(index int, v int32) {
	pos := b.offset + index*b.elemSize
	binary.LittleEndian.PutUint32(b.data[pos:], uint32(v))
}

func (b typedBuf) asElemSize(elemSize int) typedBuf {
	return typedBuf{data: b.data, offset: b.offset, elemSize: elemSize}
}

// AcmDecoder decodes ACM compressed audio to raw 16-bit PCM samples.
type AcmDecoder struct {
	numSamples  int
	numChannels int
	sampleRate  int
	levels      int
	subBlocks   int
	blockSize   int

	samplesReady int
	samplesLeft  int
	block        typedBuf // blockSize int32 elements
	values       typedBuf // current decoded block
	valuesOffset int

	unpacker *valueUnpacker
	decoder  *subbandDecoder
}

// NewAcmDecoder creates a decoder from the parsed ACM header fields and compressed data bytes.
func NewAcmDecoder(numSamples int, numChannels int, sampleRate int, levels int, subBlocks int, compressedData []byte) (*AcmDecoder, error) {
	if numSamples < 0 {
		return nil, fmt.Errorf("invalid number of samples: %d", numSamples)
	}
	if numChannels < 1 || numChannels > 2 {
		return nil, fmt.Errorf("unsupported number of channels: %d", numChannels)
	}
	if sampleRate < 4096 || sampleRate > 192000 {
		return nil, fmt.Errorf("unsupported sample rate: %d", sampleRate)
	}

	blockSize := (1 << levels) * subBlocks

	d := &AcmDecoder{
		numSamples:  numSamples,
		numChannels: numChannels,
		sampleRate:  sampleRate,
		levels:      levels,
		subBlocks:   subBlocks,
		blockSize:   blockSize,
		samplesLeft: numSamples,
		block:       allocTypedBuf(blockSize, 4),
	}

	bufB := newTypedBuf(compressedData, 0, 1)
	d.unpacker = newValueUnpacker(levels, subBlocks, bufB)
	d.decoder = newSubbandDecoder(levels)

	return d, nil
}

// ReadSamples decodes up to sampleCount samples as 16-bit LE PCM into outBuffer.
func (d *AcmDecoder) ReadSamples(outBuffer []byte, sampleCount int) int {
	pos := 0
	res := 0

	for res < sampleCount {
		if d.samplesReady == 0 {
			if d.samplesLeft == 0 {
				break
			}
			d.makeNewSamples()
		}
		val := int16(d.values.getInt32(0) >> d.levels)
		binary.LittleEndian.PutUint16(outBuffer[pos:], uint16(val))
		d.values.addOffset(1)
		pos += 2
		res++
		d.samplesReady--
	}

	// fill remaining with silence
	for i := res; i < sampleCount; i++ {
		binary.LittleEndian.PutUint16(outBuffer[pos:], 0)
		pos += 2
	}

	return res
}

func (d *AcmDecoder) makeNewSamples() {
	d.unpacker.getOneBlock(d.block)
	d.decoder.decode(d.block, d.subBlocks)
	d.values = d.block.clone()
	ready := d.blockSize
	if d.samplesLeft < ready {
		ready = d.samplesLeft
	}
	d.samplesReady = ready
	d.samplesLeft -= ready
}

// valueUnpacker reads compressed values from the ACM bitstream.
type valueUnpacker struct {
	levels    int
	subBlocks int
	sbSize    int // 1 << levels
	bufferB   typedBuf
	nextBits  int
	availBits int
	ampBuf    typedBuf // 0x10000 int16 elements
	block     typedBuf // reference to current block being filled
}

func newValueUnpacker(levels, subBlocks int, bufB typedBuf) *valueUnpacker {
	u := &valueUnpacker{
		levels:    levels,
		subBlocks: subBlocks,
		sbSize:    1 << levels,
		bufferB:   bufB.clone(),
	}

	u.ampBuf = allocTypedBuf(0x10000, 2)
	return u
}

// middle reads/writes the amplitude buffer relative to the center (offset 0x8000).
func (u *valueUnpacker) middleGet(index int) int16 {
	return u.ampBuf.getInt16(0x8000 + index)
}

func (u *valueUnpacker) middlePut(index int, v int16) {
	u.ampBuf.putInt16(0x8000+index, v)
}

func (u *valueUnpacker) prepareBits(bits int) {
	for bits > u.availBits {
		var oneByte int
		if u.bufferB.byteOffset() < u.bufferB.byteLen() {
			oneByte = int(u.bufferB.getByte(0)) & 0xff
			u.bufferB.addOffset(1)
		}
		u.nextBits |= oneByte << u.availBits
		u.availBits += 8
	}
}

func (u *valueUnpacker) getBits(bits int) int {
	u.prepareBits(bits)
	res := u.nextBits
	u.availBits -= bits
	u.nextBits >>= bits
	return res
}

func (u *valueUnpacker) getOneBlock(blockI typedBuf) {
	u.block = blockI.clone()
	pwr := u.getBits(4) & 0x0f
	val := u.getBits(16) & 0xffff
	count := 1 << pwr
	v := 0

	for i := 0; i < count; i++ {
		u.middlePut(i, int16(v))
		v += val
	}
	v = -val
	for i := 0; i < count; i++ {
		u.middlePut(-i-1, int16(v))
		v -= val
	}

	for pass := 0; pass < u.sbSize; pass++ {
		idx := u.getBits(5) & 0x1f
		if u.fillerProc(idx, pass) == 0 {
			return
		}
	}
}

func (u *valueUnpacker) fillerProc(fn, pass int) int {
	switch fn & 31 {
	case 0:
		return u.zeroFill(pass)
	case 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16:
		return u.linearFill(pass, fn)
	case 17:
		return u.k1Bits3(pass)
	case 18:
		return u.k1Bits2(pass)
	case 19:
		return u.t1Bits5(pass)
	case 20:
		return u.k2Bits4(pass)
	case 21:
		return u.k2Bits3(pass)
	case 22:
		return u.t2Bits7(pass)
	case 23:
		return u.k3Bits5(pass)
	case 24:
		return u.k3Bits4(pass)
	case 26:
		return u.k4Bits5(pass)
	case 27:
		return u.k4Bits4(pass)
	case 29:
		return u.t3Bits7(pass)
	default:
		return 0
	}
}

func (u *valueUnpacker) zeroFill(pass int) int {
	step := u.sbSize
	for i := 0; i < u.subBlocks; i++ {
		u.block.putInt32(i*step+pass, 0)
	}
	return 1
}

func (u *valueUnpacker) linearFill(pass, idx int) int {
	mask := (1 << idx) - 1
	lbOfs := -(1 << (idx - 1))
	for i := 0; i < u.subBlocks; i++ {
		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(lbOfs+(u.getBits(idx)&mask))))
	}
	return 1
}

func (u *valueUnpacker) k1Bits3(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(3)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
			i++
			if i == u.subBlocks {
				break
			}
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else if (u.nextBits & 2) == 0 {
			u.availBits -= 2
			u.nextBits >>= 2
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else {
			idx := 1
			if (u.nextBits & 4) == 0 {
				idx = -1
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(idx)))
			u.availBits -= 3
			u.nextBits >>= 3
		}
	}
	return 1
}

func (u *valueUnpacker) k1Bits2(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(2)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else {
			idx := 1
			if (u.nextBits & 2) == 0 {
				idx = -1
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(idx)))
			u.availBits -= 2
			u.nextBits >>= 2
		}
	}
	return 1
}

func (u *valueUnpacker) t1Bits5(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		val := byte(u.getBits(5) & 0x1f)
		val = table1[val]

		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val&3)-1)))
		i++
		if i == u.subBlocks {
			break
		}
		val >>= 2
		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val&3)-1)))
		i++
		if i == u.subBlocks {
			break
		}
		val >>= 2
		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val)-1)))
	}
	return 1
}

func (u *valueUnpacker) k2Bits4(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(4)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
			i++
			if i == u.subBlocks {
				break
			}
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else if (u.nextBits & 2) == 0 {
			u.availBits -= 2
			u.nextBits >>= 2
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else {
			var v int16
			if (u.nextBits & 8) != 0 {
				if (u.nextBits & 4) != 0 {
					v = u.middleGet(2)
				} else {
					v = u.middleGet(1)
				}
			} else {
				if (u.nextBits & 4) != 0 {
					v = u.middleGet(-1)
				} else {
					v = u.middleGet(-2)
				}
			}
			u.block.putInt32(i*u.sbSize+pass, int32(v))
			u.availBits -= 4
			u.nextBits >>= 4
		}
	}
	return 1
}

func (u *valueUnpacker) k2Bits3(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(3)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else {
			var v int16
			if (u.nextBits & 4) != 0 {
				if (u.nextBits & 2) != 0 {
					v = u.middleGet(2)
				} else {
					v = u.middleGet(1)
				}
			} else {
				if (u.nextBits & 2) != 0 {
					v = u.middleGet(-1)
				} else {
					v = u.middleGet(-2)
				}
			}
			u.block.putInt32(i*u.sbSize+pass, int32(v))
			u.availBits -= 3
			u.nextBits >>= 3
		}
	}
	return 1
}

func (u *valueUnpacker) t2Bits7(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		val := int16(u.getBits(7) & 0x7f)
		val = table2[val]

		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val&7)-2)))
		i++
		if i == u.subBlocks {
			break
		}
		val >>= 3
		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val&7)-2)))
		i++
		if i == u.subBlocks {
			break
		}
		val >>= 3
		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val)-2)))
	}
	return 1
}

func (u *valueUnpacker) k3Bits5(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(5)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
			i++
			if i == u.subBlocks {
				break
			}
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else if (u.nextBits & 2) == 0 {
			u.availBits -= 2
			u.nextBits >>= 2
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else if (u.nextBits & 4) == 0 {
			idx := 1
			if (u.nextBits & 8) == 0 {
				idx = -1
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(idx)))
			u.availBits -= 4
			u.nextBits >>= 4
		} else {
			u.availBits -= 5
			val := (u.nextBits & 0x18) >> 3
			u.nextBits >>= 5
			if val >= 2 {
				val += 3
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(val-3)))
		}
	}
	return 1
}

func (u *valueUnpacker) k3Bits4(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(4)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else if (u.nextBits & 2) == 0 {
			u.availBits -= 3
			idx := 1
			if (u.nextBits & 4) == 0 {
				idx = -1
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(idx)))
			u.nextBits >>= 3
		} else {
			val := (u.nextBits & 0x0c) >> 2
			u.availBits -= 4
			u.nextBits >>= 4
			if val >= 2 {
				val += 3
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(val-3)))
		}
	}
	return 1
}

func (u *valueUnpacker) k4Bits5(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(5)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
			i++
			if i == u.subBlocks {
				break
			}
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else if (u.nextBits & 2) == 0 {
			u.availBits -= 2
			u.nextBits >>= 2
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else {
			val := (u.nextBits & 0x1c) >> 2
			if val >= 4 {
				val++
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(val-4)))
			u.availBits -= 5
			u.nextBits >>= 5
		}
	}
	return 1
}

func (u *valueUnpacker) k4Bits4(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		u.prepareBits(4)
		if (u.nextBits & 1) == 0 {
			u.availBits--
			u.nextBits >>= 1
			u.block.putInt32(i*u.sbSize+pass, 0)
		} else {
			val := (u.nextBits & 0x0e) >> 1
			u.availBits -= 4
			u.nextBits >>= 4
			if val >= 4 {
				val++
			}
			u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(val-4)))
		}
	}
	return 1
}

func (u *valueUnpacker) t3Bits7(pass int) int {
	for i := 0; i < u.subBlocks; i++ {
		val := int16(u.getBits(7) & 0x7f)
		val = table3[val]

		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val&0x0f)-5)))
		i++
		if i == u.subBlocks {
			break
		}
		val >>= 4
		u.block.putInt32(i*u.sbSize+pass, int32(u.middleGet(int(val)-5)))
	}
	return 1
}

// subbandDecoder performs inverse subband filtering to reconstruct PCM samples.
type subbandDecoder struct {
	levels    int
	blockSize int // 1 << levels
	memBuf    typedBuf
}

func newSubbandDecoder(levels int) *subbandDecoder {
	d := &subbandDecoder{
		levels:    levels,
		blockSize: 1 << levels,
	}

	memSize := 0
	if levels > 0 {
		memSize = 3*(d.blockSize>>1) - 2
	}
	if memSize > 0 {
		// First (blockSize>>1) elements are int16 (2 bytes), rest are int32 (4 bytes).
		// Total byte size: (blockSize>>1)*2 + ((blockSize>>1)-2)*2 * 4
		// Simplify: just allocate enough for all as int32 which covers both views.
		d.memBuf = allocTypedBuf(memSize, 4)
	}

	return d
}

func (d *subbandDecoder) decode(bufferI typedBuf, blocks int) {
	if d.levels == 0 {
		return
	}

	bufI := bufferI.clone()
	memI := d.memBuf.clone()
	sbSize := d.blockSize >> 1

	blocks <<= 1
	d.sub4d3fcc(memI.asElemSize(2), bufI, sbSize, blocks)
	memI.addOffset(sbSize)

	for i := 0; i < blocks; i++ {
		bufI.putInt32(i*sbSize, bufI.getInt32(i*sbSize)+1)
	}

	sbSize >>= 1
	blocks <<= 1

	for sbSize != 0 {
		d.sub4d420c(memI, bufI, sbSize, blocks)
		memI.offset += sbSize * 2 * 4 // advance by sbSize*2 int32 elements
		sbSize >>= 1
		blocks <<= 1
	}
}

func (d *subbandDecoder) sub4d3fcc(memoryS typedBuf, bufferI typedBuf, sbSize, blocks int) {
	memS := memoryS.clone()
	bufI := bufferI.clone()
	var row0, row1, row2, row3, db0, db1 int32
	sbSize2 := sbSize * 2
	sbSize3 := sbSize * 3

	if blocks == 2 {
		for i := 0; i < sbSize; i++ {
			row0 = bufI.getInt32(0)
			row1 = bufI.getInt32(sbSize)
			bufI.putInt32(0, bufI.getInt32(0)+int32(memS.getInt16(0))+(int32(memS.getInt16(1))<<1))
			bufI.putInt32(sbSize, (row0<<1)-int32(memS.getInt16(1))-bufI.getInt32(sbSize))
			memS.putInt16(0, int16(row0))
			memS.putInt16(1, int16(row1))

			memS.addOffset(2)
			bufI.addOffset(1)
		}
	} else if blocks == 4 {
		for i := 0; i < sbSize; i++ {
			row0 = bufI.getInt32(0)
			row1 = bufI.getInt32(sbSize)
			row2 = bufI.getInt32(sbSize2)
			row3 = bufI.getInt32(sbSize3)

			bufI.putInt32(0, int32(memS.getInt16(0))+(int32(memS.getInt16(1))<<1)+row0)
			bufI.putInt32(sbSize, -int32(memS.getInt16(1))+(row0<<1)-row1)
			bufI.putInt32(sbSize2, row0+(row1<<1)+row2)
			bufI.putInt32(sbSize3, -row1+(row2<<1)-row3)

			memS.putInt16(0, int16(row2))
			memS.putInt16(1, int16(row3))

			memS.addOffset(2)
			bufI.addOffset(1)
		}
	} else {
		buf2I := bufI.clone()
		for i := 0; i < sbSize; i++ {
			buf2I.setOffset(bufI.byteOffset())
			if (blocks & 2) != 0 {
				row0 = buf2I.getInt32(0)
				row1 = buf2I.getInt32(sbSize)

				buf2I.putInt32(0, int32(memS.getInt16(0))+(int32(memS.getInt16(1))<<1)+row0)
				buf2I.putInt32(sbSize, -int32(memS.getInt16(1))+(row0<<1)-row1)
				buf2I.offset += sbSize2 * buf2I.elemSize

				db0 = row0
				db1 = row1
			} else {
				db0 = int32(memS.getInt16(0))
				db1 = int32(memS.getInt16(1))
			}

			for j := 0; j < (blocks >> 2); j++ {
				row0 = buf2I.getInt32(0)
				buf2I.putInt32(0, db0+(db1<<1)+row0)
				buf2I.offset += sbSize * buf2I.elemSize

				row1 = buf2I.getInt32(0)
				buf2I.putInt32(0, -db1+(row0<<1)-row1)
				buf2I.offset += sbSize * buf2I.elemSize

				row2 = buf2I.getInt32(0)
				buf2I.putInt32(0, row0+(row1<<1)+row2)
				buf2I.offset += sbSize * buf2I.elemSize

				row3 = buf2I.getInt32(0)
				buf2I.putInt32(0, -row1+(row2<<1)-row3)
				buf2I.offset += sbSize * buf2I.elemSize

				db0 = row2
				db1 = row3
			}
			memS.putInt16(0, int16(row2))
			memS.putInt16(1, int16(row3))
			memS.addOffset(2)
			bufI.addOffset(1)
		}
	}
}

func (d *subbandDecoder) sub4d420c(memoryI typedBuf, bufferI typedBuf, sbSize, blocks int) {
	memI := memoryI.clone()
	bufI := bufferI.clone()
	var row0, row1, row2, row3, db0, db1 int32
	sbSize2 := sbSize * 2
	sbSize3 := sbSize * 3

	if blocks == 4 {
		for i := 0; i < sbSize; i++ {
			row0 = bufI.getInt32(0)
			row1 = bufI.getInt32(sbSize)
			row2 = bufI.getInt32(sbSize2)
			row3 = bufI.getInt32(sbSize3)

			bufI.putInt32(0, memI.getInt32(0)+(memI.getInt32(1)<<1)+row0)
			bufI.putInt32(sbSize, -memI.getInt32(1)+(row0<<1)-row1)
			bufI.putInt32(sbSize2, row0+(row1<<1)+row2)
			bufI.putInt32(sbSize3, -row1+(row2<<1)-row3)

			memI.putInt32(0, row2)
			memI.putInt32(1, row3)

			memI.addOffset(2)
			bufI.addOffset(1)
		}
	} else {
		buf2I := bufI.clone()
		for i := 0; i < sbSize; i++ {
			buf2I.setOffset(bufI.byteOffset())
			db0 = memI.getInt32(0)
			db1 = memI.getInt32(1)
			for j := 0; j < (blocks >> 2); j++ {
				row0 = buf2I.getInt32(0)
				buf2I.putInt32(0, db0+(db1<<1)+row0)
				buf2I.offset += sbSize * buf2I.elemSize

				row1 = buf2I.getInt32(0)
				buf2I.putInt32(0, -db1+(row0<<1)-row1)
				buf2I.offset += sbSize * buf2I.elemSize

				row2 = buf2I.getInt32(0)
				buf2I.putInt32(0, row0+(row1<<1)+row2)
				buf2I.offset += sbSize * buf2I.elemSize

				row3 = buf2I.getInt32(0)
				buf2I.putInt32(0, -row1+(row2<<1)-row3)
				buf2I.offset += sbSize * buf2I.elemSize

				db0 = row2
				db1 = row3
			}
			memI.putInt32(0, row2)
			memI.putInt32(1, row3)

			memI.addOffset(2)
			bufI.addOffset(1)
		}
	}
}

// SPDX-SnippetEnd
