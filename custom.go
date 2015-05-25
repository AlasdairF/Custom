package custom

import (
 "compress/zlib"
 "io"
 "os"
 "math"
 "errors"
)

type Reader struct {
 f io.ReadCloser
 at int		// the cursor for where I am in buf
 n int		// how much uncompressed but as of yet unparsed data is left in buf
 buf []byte	// the buffer for reading data
}

type Writer struct {
 f *zlib.Writer
 buf8 []byte
 buf16 []byte
 buf24 []byte
 buf32 []byte
 buf48 []byte
 buf64 []byte
}

func (w *Writer) Close() {
	w.f.Close()
}

func (r *Reader) Close() {
	r.f.Close()
}

func (r *Reader) EOF() error {
	_, err := r.f.Read(r.buf)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return errors.New(`Not EOF`)
	}
	return err
}

func NewReader(f io.Reader, buffersize int) (*Reader, error) {
	r := new(Reader)
	var err error
	r.f, err = zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	r.buf = make([]byte, buffersize)
	return r, nil
}

func NewWriter(f io.Writer) *Writer {
	w := new(Writer)
	w.f = zlib.NewWriter(f)
	w.buf8 = make([]byte, 1)
	w.buf16 = make([]byte, 2)
	w.buf24 = make([]byte, 3)
	w.buf32 = make([]byte, 4)
	w.buf48 = make([]byte, 6)
	w.buf64 = make([]byte, 8)
	return w
}

func NewWriterLevel(f io.Writer, level int) *Writer {
	w := new(Writer)
	w.f, _ = zlib.NewWriterLevel(f, level)
	w.buf8 = make([]byte, 1)
	w.buf16 = make([]byte, 2)
	w.buf24 = make([]byte, 3)
	w.buf32 = make([]byte, 4)
	w.buf48 = make([]byte, 6)
	w.buf64 = make([]byte, 8)
	return w
}

func (w *Writer) Write(b []byte) {
	w.f.Write(b)
}

func (w *Writer) WriteBool2(v1, v2 bool) {
	if v1 {
		w.buf8[0] = 1
	} else {
		w.buf8[0] = 0
	}
	if v2 {
		w.buf8[0] |= 2
	}
	w.f.Write(w.buf8)
}

func (w *Writer) WriteBool(v bool) {
	if v {
		w.buf8[0] = 1
	} else {
		w.buf8[0] = 0
	}
	w.f.Write(w.buf8)
}

func (w *Writer) Write4(v1, v2 uint8) {
	v1 |= v2 << 4
	w.buf8[0] = v1
	w.f.Write(w.buf8)
}

func (w *Writer) Write8(v uint8) {
	w.buf8[0] = v
	w.f.Write(w.buf8)
}

func (w *Writer) Write16(v uint16) {
	w.buf16[0] = byte(v)
	w.buf16[1] = byte(v >> 8)
	w.f.Write(w.buf16)
}

// If it's less than 255 then it's encoded in the 1st byte, otherwise 1st byte is 255 and it's encoded in two more bytes
// This is only useful if it is expected that the value will be <255 more than half the time
func (w *Writer) Write16Variable(v uint16) {
	if v < 255 {
		w.buf8[0] = byte(v)
		w.f.Write(w.buf8)
		return
	}
	w.buf24[0] = 255
	w.buf24[1] = byte(v)
	w.buf24[2] = byte(v >> 8)
	w.f.Write(w.buf24)
}

func (w *Writer) WriteInt16Variable(v int16) {
	if v > -128 && v < 128 {
		w.buf8[0] = byte(v + 127)
		w.f.Write(w.buf8)
		return
	}
	v2 := uint16(v)
	w.buf24[0] = 255
	w.buf24[1] = byte(v2)
	w.buf24[2] = byte(v2 >> 8)
	w.f.Write(w.buf24)
}

func (w *Writer) Write24(v uint32) {
	w.buf24[0] = byte(v)
	w.buf24[1] = byte(v >> 8)
	w.buf24[2] = byte(v >> 16)
	w.f.Write(w.buf24)
}

func (w *Writer) Write32(v uint32) {
	w.buf32[0] = byte(v)
	w.buf32[1] = byte(v >> 8)
	w.buf32[2] = byte(v >> 16)
	w.buf32[3] = byte(v >> 24)
	w.f.Write(w.buf32)
}

func (w *Writer) Write48(v uint64) {
	w.buf48[0] = byte(v)
	w.buf48[1] = byte(v >> 8)
	w.buf48[2] = byte(v >> 16)
	w.buf48[3] = byte(v >> 24)
	w.buf48[4] = byte(v >> 32)
	w.buf48[5] = byte(v >> 40)
	w.f.Write(w.buf48)
}

func (w *Writer) WriteFloat32(flt float32) {
	w.Write32(math.Float32bits(flt))
}

func (w *Writer) Write64(v uint64) {
	w.buf64[0] = byte(v)
	w.buf64[1] = byte(v >> 8)
	w.buf64[2] = byte(v >> 16)
	w.buf64[3] = byte(v >> 24)
	w.buf64[4] = byte(v >> 32)
	w.buf64[5] = byte(v >> 40)
	w.buf64[6] = byte(v >> 48)
	w.buf64[7] = byte(v >> 56)
	w.f.Write(w.buf64)
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (w *Writer) Write64Variable(v uint64) {
	s := numbytes(v)
	w.Write8(s)
	w.buf64[0] = byte(v)
	w.buf64[1] = byte(v >> 8)
	w.buf64[2] = byte(v >> 16)
	w.buf64[3] = byte(v >> 24)
	w.buf64[4] = byte(v >> 32)
	w.buf64[5] = byte(v >> 40)
	w.buf64[6] = byte(v >> 48)
	w.buf64[7] = byte(v >> 56)
	w.f.Write(w.buf64[0:s])
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (w *Writer) Write64Variable2(v1 uint64, v2 uint64) {
	s1 := numbytes(v1)
	s2 := numbytes(v2)
	w.Write8((s1 << 4) | s2)
	w.buf64[0] = byte(v1)
	w.buf64[1] = byte(v1 >> 8)
	w.buf64[2] = byte(v1 >> 16)
	w.buf64[3] = byte(v1 >> 24)
	w.buf64[4] = byte(v1 >> 32)
	w.buf64[5] = byte(v1 >> 40)
	w.buf64[6] = byte(v1 >> 48)
	w.buf64[7] = byte(v1 >> 56)
	w.f.Write(w.buf64[0:s1])
	w.buf64[0] = byte(v2)
	w.buf64[1] = byte(v2 >> 8)
	w.buf64[2] = byte(v2 >> 16)
	w.buf64[3] = byte(v2 >> 24)
	w.buf64[4] = byte(v2 >> 32)
	w.buf64[5] = byte(v2 >> 40)
	w.buf64[6] = byte(v2 >> 48)
	w.buf64[7] = byte(v2 >> 56)
	w.f.Write(w.buf64[0:s2])
}

func numbytes(v uint64) uint8 {
	switch {
		case v == 0: return 0
		case v < 256: return 1
		case v < 65536: return 2
		case v < 16777216: return 3
		case v < 4294967296: return 4
		case v < 1099511627776: return 5
		case v < 281474976710655: return 6
		case v < 72057594037927936: return 7
		default: return 8
	}
}

func (w *Writer) WriteFloat64(flt float64) {
	w.Write64(math.Float64bits(flt))
}

func (w *Writer) WriteString8(s string) {
	tmp := []byte(s)
	if len(tmp) > 255 {
		tmp = tmp[0:255]
	}
	w.buf8[0] = uint8(len(tmp))
	w.f.Write(w.buf8)
	w.f.Write(tmp)
}

func (w *Writer) WriteString16(s string) {
	tmp := []byte(s)
	if len(tmp) > 65535 {
		tmp = tmp[0:65535]
	}
	w.Write16(uint16(len(tmp)))
	w.f.Write(tmp)
}

func (w *Writer) WriteString32(s string) {
	tmp := []byte(s)
	if len(tmp) > 4294967295 {
		tmp = tmp[0:4294967295]
	}
	w.Write32(uint32(len(tmp)))
	w.f.Write(tmp)
}

// 12 bits and 4 bits
func (w *Writer) Write12(v1, v2 uint16) {
	v1 |= v2 << 12
	w.buf16[0] = byte(v1)
	w.buf16[1] = byte(v1 >> 8)
	w.f.Write(w.buf16)
}

func (w *Writer) WriteSpecial(v1 uint8, b1, b2, b3, b4 bool) {
	if b1 {
		v1 |= 128
	}
	if b2 {
		v1 |= 64
	}
	if b3 {
		v1 |= 32
	}
	if b4 {
		v1 |= 16
	}
	w.buf8[0] = v1
	w.f.Write(w.buf8)
}

func (w *Writer) WriteSpecial2(v1, v2, v3 uint8, b1 bool) {
	v1 |= v2 << 3
	v1 |= v3 << 5
	if b1 {
		v1 |= 128
	}
	w.buf8[0] = v1
	w.f.Write(w.buf8)
}

func (r *Reader) ReadBool2() (bool, bool) {
	for r.n == 0 {
		r.at = 0
		m, err := r.f.Read(r.buf)
		if err != nil {
			panic(err)
		}
		r.n = m
	}
	var b1, b2 bool
	switch r.buf[r.at] {
		case 1: b1 = true
		case 2: b2 = true
		case 3: b1, b2 = true, true
	}
	r.at++
	r.n--
	return b1, b2
}

func (r *Reader) ReadBool() bool {
	for r.n == 0 {
		r.at = 0
		m, err := r.f.Read(r.buf)
		if err != nil {
			panic(err)
		}
		r.n = m
	}
	var b1 bool
	if r.buf[r.at] > 0 {
		b1 = true
	}
	r.at++
	r.n--
	return b1
}

func (r *Reader) Read4() (uint8, uint8) {
	for r.n == 0 {
		r.at = 0
		m, err := r.f.Read(r.buf)
		if err != nil {
			panic(err)
		}
		r.n = m
	}
	res1, res2 := r.buf[r.at] & 15, r.buf[r.at] >> 4
	r.at++
	r.n--
	return res1, res2
}

func (r *Reader) Read8() uint8 {
	for r.n == 0 {
		r.at = 0
		m, err := r.f.Read(r.buf)
		if err != nil {
			panic(err)
		}
		r.n = m
	}
	res := r.buf[r.at]
	r.at++
	r.n--
	return res
}

func (r *Reader) Read(b []byte) {
	x := len(b)
	for r.n < x {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	copy(b, r.buf[r.at:r.at+x]) // must be copied to avoid memory leak
	r.at += x
	r.n -= x
}

func (r *Reader) Readx(x int) []byte {
	for r.n < x {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	tmp := make([]byte, x)
	copy(tmp, r.buf[r.at:r.at+x]) // must be copied to avoid memory leak
	r.at += x
	r.n -= x
	return tmp
}

func (r *Reader) ReadUTF8() []byte {
	first := r.Read8()
	if first < 128 { // length 1
		b := make([]byte, 1)
		b[0] = first
		return b
	}
	if first & 32 == 0 { // length 2
			b := make([]byte, 2)
			b[0] = first
			b[1] = r.Read8()
			return b
	} else {
		b := make([]byte, 3)
		b[0] = first
		b[1] = r.Read8()
		b[2] = r.Read8()
		return b
	}
}

func (r *Reader) Read16() uint16 {
	for r.n < 2 {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	res := uint16(r.buf[r.at]) | uint16(r.buf[r.at+1])<<8
	r.at += 2
	r.n -= 2
	return res
}

// If it's less than 255 then it's encoded in the 1st byte, otherwise 1st byte is 255 and it's encoded in two more bytes
// This is only useful if it is expected that the value will be <255 more than half the time
func (r *Reader) Read16Variable() uint16 {
	v := r.Read8()
	if v < 255 {
		return uint16(v)
	}
	return r.Read16()
}

func (r *Reader) ReadInt16Variable() int16 {
	v := r.Read8()
	if v < 255 {
		return int16(v) - 127
	}
	return int16(r.Read16())
}

func (r *Reader) Read24() uint32 {
	for r.n < 3 {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	res := uint32(r.buf[r.at]) | uint32(r.buf[r.at+1])<<8 | uint32(r.buf[r.at+2])<<16
	r.at += 3
	r.n -= 3
	return res
}

func (r *Reader) Read32() uint32 {
	for r.n < 4 {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	res := uint32(r.buf[r.at]) | uint32(r.buf[r.at+1])<<8 | uint32(r.buf[r.at+2])<<16 | uint32(r.buf[r.at+3])<<24
	r.at += 4
	r.n -= 4
	return res
}

func (r *Reader) ReadFloat32() float32 {
	return math.Float32frombits(r.Read32())
}

func (r *Reader) Read48() uint64 {
	for r.n < 6 {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	res := uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40
	r.at += 6
	r.n -= 6
	return res
}

func (r *Reader) Read64() uint64 {
	for r.n < 8 {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	res := uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40 | uint64(r.buf[r.at+6])<<48 | uint64(r.buf[r.at+7])<<56
	r.at += 8
	r.n -= 8
	return res
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (r *Reader) Read64Variable() uint64 {
	s1 := int(r.Read8())
	for r.n < s1 {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	var res1 uint64
	switch s1 {
		case 1: res1 = uint64(r.buf[r.at])
		case 2: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8
		case 3: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16
		case 4: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24
		case 5: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32
		case 6: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40
		case 7: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40 | uint64(r.buf[r.at+6])<<48
		case 8: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40 | uint64(r.buf[r.at+6])<<48 | uint64(r.buf[r.at+7])<<56
	}
	r.at += s1
	r.n -= s1
	return res1
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (r *Reader) Read64Variable2() (uint64, uint64) {
	s2 := r.Read8()
	s1 := s2 >> 4
	s2 &= 15
	for r.n < int(s1 + s2) {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	var res1, res2 uint64
	switch s1 {
		case 1: res1 = uint64(r.buf[r.at])
		case 2: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8
		case 3: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16
		case 4: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24
		case 5: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32
		case 6: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40
		case 7: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40 | uint64(r.buf[r.at+6])<<48
		case 8: res1 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40 | uint64(r.buf[r.at+6])<<48 | uint64(r.buf[r.at+7])<<56
	}
	r.at += int(s1)
	r.n -= int(s1)
	switch s2 {
		case 1: res2 = uint64(r.buf[r.at])
		case 2: res2 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8
		case 3: res2 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16
		case 4: res2 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24
		case 5: res2 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32
		case 6: res2 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40
		case 7: res2 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40 | uint64(r.buf[r.at+6])<<48
		case 8: res2 = uint64(r.buf[r.at]) | uint64(r.buf[r.at+1])<<8 | uint64(r.buf[r.at+2])<<16 | uint64(r.buf[r.at+3])<<24 | uint64(r.buf[r.at+4])<<32 | uint64(r.buf[r.at+5])<<40 | uint64(r.buf[r.at+6])<<48 | uint64(r.buf[r.at+7])<<56
	}
	r.at += int(s2)
	r.n -= int(s2)
	return res1, res2
}

func (r *Reader) ReadFloat64() float64 {
	return math.Float64frombits(r.Read64())
}

func (r *Reader) ReadString8() string {
	return string(r.Readx(int(r.Read8())))
}

func (r *Reader) ReadString16() string {
	return string(r.Readx(int(r.Read16())))
}

func (r *Reader) ReadString32() string {
	return string(r.Readx(int(r.Read32())))
}

// 12 bits for uint16 and 4 bits for uint8
func (r *Reader) Read12() (uint16, uint16) {
	for r.n < 2 {
		copy(r.buf, r.buf[r.at:r.at+r.n])
		r.at = 0
		m, err := r.f.Read(r.buf[r.n:])
		if err != nil {
			panic(err)
		}
		r.n += m
	}
	res := uint16(r.buf[r.at]) | uint16(r.buf[r.at+1])<<8
	r.at += 2
	r.n -= 2
	return res & 4095, res >> 12
}

func (r *Reader) ReadSpecial() (uint8, bool, bool, bool, bool) {
	for r.n == 0 {
		r.at = 0
		m, err := r.f.Read(r.buf)
		if err != nil {
			panic(err)
		}
		r.n = m
	}
	c := r.buf[r.at]
	var b1, b2, b3, b4 bool
	if c & 128 > 0 {
		b1 = true
	}
	if c & 64 > 0 {
		b2 = true
	}
	if c & 32 > 0 {
		b3 = true
	}
	if c & 16 > 0 {
		b4 = true
	}
	r.at++
	r.n--
	return c & 7, b1, b2, b3, b4
}

func (r *Reader) ReadSpecial2() (uint8, uint8, uint8, bool) {
	for r.n == 0 {
		r.at = 0
		m, err := r.f.Read(r.buf)
		if err != nil {
			panic(err)
		}
		r.n = m
	}
	c := r.buf[r.at]
	var b1 bool
	if c & 128 > 0 {
		b1 = true
	}
	v1 := c & 7
	v2 := (c >> 3) & 3
	v3 := (c >> 5) & 3
	r.at++
	r.n--
	return v1, v2, v3, b1
}
