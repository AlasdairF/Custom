package custom

import (
 "github.com/AlasdairF/Buffer"
 "compress/zlib"
 "io"
 "math"
 "errors"
)

type Reader struct {
	f io.ReadCloser
	at int		// the cursor for where I am in buf
	n int		// how much uncompressed but as of yet unparsed data is left in buf
	buf []byte	// the buffer for reading data
}

type BytesReader struct {
	data []byte
	cursor int
}

type Writer struct {
	f *buffer.Writer
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

func (r *BytesReader) EOF() error {
	if r.cursor == len(r.data) {
		return nil
	}
	return errors.New(`Not EOF`)
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

func NewBytesReader(data []byte) (*BytesReader, error) {
	return &BytesReader{data: data}, nil
}

func LoadBytes(r io.Reader) ([]byte, error) {
	z, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	b, err := ioutil.ReadAll(z)
	return b, err
}

func NewWriter(f io.Writer) *Writer {
	return &Writer{f: buffer.NewWriter(zlib.NewWriter(f))}
}

func NewWriterLevel(f io.Writer, level int) *Writer {
	w := new(Writer)
	z, _ := zlib.NewWriterLevel(f, level)
	w.f = buffer.NewWriter(z)
	return w
}

// -------- WRITING --------

func (w *Writer) Write(b []byte) {
	w.f.Write(b)
}

func (w *Writer) WriteBool2(v1, v2 bool) {
	var b byte
	if v1 {
		b = 1
	}
	if v2 {
		b |= 2
	}
	w.f.WriteByte(b)
}

func (w *Writer) WriteBool(v bool) {
	var b byte
	if v {
		b = 1
	}
	w.f.WriteByte(b)
}

func (w *Writer) Write4(v1, v2 uint8) {
	v1 |= v2 << 4
	w.f.WriteByte(v1)
}

func (w *Writer) Write8(v uint8) {
	w.f.WriteByte(v)
}

func (w *Writer) Write16(v uint16) {
	w.f.Write2Bytes(byte(v), byte(v >> 8))
}

// If it's less than 255 then it's encoded in the 1st byte, otherwise 1st byte is 255 and it's encoded in two more bytes
// This is only useful if it is expected that the value will be <255 more than half the time
func (w *Writer) Write16Variable(v uint16) {
	if v < 255 {
		w.f.WriteByte(byte(v))
		return
	}
	w.f.Write3Bytes(255, byte(v), byte(v >> 8))
}

func (w *Writer) WriteInt16Variable(v int16) {
	if v > -128 && v < 128 {
		w.f.WriteByte(byte(v + 127))
		return
	}
	v2 := uint16(v)
	w.f.Write3Bytes(255, byte(v2), byte(v2 >> 8))
}

func (w *Writer) Write24(v uint32) {
	w.f.Write3Bytes(byte(v), byte(v >> 8), byte(v >> 16))
}

func (w *Writer) Write32(v uint32) {
	w.f.Write4Bytes(byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24))
}

func (w *Writer) Write48(v uint64) {
	w.f.Write6Bytes(byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32), byte(v >> 40))
}

func (w *Writer) WriteFloat32(flt float32) {
	w.Write32(math.Float32bits(flt))
}

func (w *Writer) Write64(v uint64) {
	w.f.Write8Bytes(byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56))
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (w *Writer) Write64Variable(v uint64) {
	switch numbytes(v) {
		case 0: w.f.WriteByte(0)
		case 1: w.f.Write2Bytes(1, byte(v))
		case 2: w.f.Write3Bytes(2, byte(v), byte(v >> 8))
		case 3: w.f.Write4Bytes(3, byte(v), byte(v >> 8), byte(v >> 16))
		case 4: w.f.Write5Bytes(4, byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24))
		case 5: w.f.Write6Bytes(5, byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32))
		case 6: w.f.Write7Bytes(6, byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32), byte(v >> 40))
		case 7: w.f.Write8Bytes(7, byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24), byte(v >> 32), byte(v >> 40), byte(v >> 48))
		case 8: w.f.Write9Bytes(8, byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 25), byte(v >> 32), byte(v >> 40), byte(v >> 48), byte(v >> 56))
	}
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (w *Writer) Write64Variable2(v1 uint64, v2 uint64) {
	s1 := numbytes(v1)
	s2 := numbytes(v2)
	w.f.WriteByte((s1 << 4) | s2)
	switch s1 {
		case 1: w.f.WriteByte(byte(v1))
		case 2: w.f.Write2Bytes(byte(v1), byte(v1 >> 8))
		case 3: w.f.Write3Bytes(byte(v1), byte(v1 >> 8), byte(v1 >> 16))
		case 4: w.f.Write4Bytes(byte(v1), byte(v1 >> 8), byte(v1 >> 16), byte(v1 >> 24))
		case 5: w.f.Write5Bytes(byte(v1), byte(v1 >> 8), byte(v1 >> 16), byte(v1 >> 24), byte(v1 >> 32))
		case 6: w.f.Write6Bytes(byte(v1), byte(v1 >> 8), byte(v1 >> 16), byte(v1 >> 24), byte(v1 >> 32), byte(v1 >> 40))
		case 7: w.f.Write7Bytes(byte(v1), byte(v1 >> 8), byte(v1 >> 16), byte(v1 >> 24), byte(v1 >> 32), byte(v1 >> 40), byte(v1 >> 48))
		case 8: w.f.Write8Bytes(byte(v1), byte(v1 >> 8), byte(v1 >> 16), byte(v1 >> 25), byte(v1 >> 32), byte(v1 >> 40), byte(v1 >> 48), byte(v1 >> 56))
	}
	switch s2 {
		case 1: w.f.WriteByte(byte(v2))
		case 2: w.f.Write2Bytes(byte(v2), byte(v2 >> 8))
		case 3: w.f.Write3Bytes(byte(v2), byte(v2 >> 8), byte(v2 >> 16))
		case 4: w.f.Write4Bytes(byte(v2), byte(v2 >> 8), byte(v2 >> 16), byte(v2 >> 24))
		case 5: w.f.Write5Bytes(byte(v2), byte(v2 >> 8), byte(v2 >> 16), byte(v2 >> 24), byte(v2 >> 32))
		case 6: w.f.Write6Bytes(byte(v2), byte(v2 >> 8), byte(v2 >> 16), byte(v2 >> 24), byte(v2 >> 32), byte(v2 >> 40))
		case 7: w.f.Write7Bytes(byte(v2), byte(v2 >> 8), byte(v2 >> 16), byte(v2 >> 24), byte(v2 >> 32), byte(v2 >> 40), byte(v2 >> 48))
		case 8: w.f.Write8Bytes(byte(v2), byte(v2 >> 8), byte(v2 >> 16), byte(v2 >> 25), byte(v2 >> 32), byte(v2 >> 40), byte(v2 >> 48), byte(v2 >> 56))
	}
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
	l := len(s)
	w.f.WriteByte(uint8(l))
	if l > 255 {
		w.f.WriteString(s[0:255])
	} else {
		w.f.WriteString(s)
	}
}

func (w *Writer) WriteString16(s string) {
	l := len(s)
	w.Write16(uint16(l))
	if l > 65535 {
		w.f.WriteString(s[0:65535])
	} else {
		w.f.WriteString(s)
	}
}

func (w *Writer) WriteString32(s string) {
	l := len(s)
	w.Write32(uint32(l))
	if l > 4294967295 {
		w.f.WriteString(s[0:4294967295])
	} else {
		w.f.WriteString(s)
	}
}

// 12 bits and 4 bits
func (w *Writer) Write12(v1, v2 uint16) {
	v1 |= v2 << 12
	w.f.Write2Bytes(byte(v1), byte(v1 >> 8))
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
	w.f.WriteByte(v1)
}

func (w *Writer) WriteSpecial2(v1, v2, v3 uint8, b1 bool) {
	v1 |= v2 << 3
	v1 |= v3 << 5
	if b1 {
		v1 |= 128
	}
	w.f.WriteByte(v1)
}

// -------- READING --------

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
		return []byte{first}
	}
	if first & 32 == 0 { // length 2
			return []byte{first, r.Read8()}
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

// -------- BytesReader READING --------

func (r *BytesReader) ReadBool2() (bool, bool) {
	var b1, b2 bool
	switch r.data[r.cursor] {
		case 1: b1 = true
		case 2: b2 = true
		case 3: b1, b2 = true, true
	}
	r.cursor++
	return b1, b2
}

func (r *BytesReader) ReadBool() bool {
	var b1 bool
	if r.data[r.cursor] > 0 {
		b1 = true
	}
	r.cursor++
	return b1
}

func (r *BytesReader) Read4() (uint8, uint8) {
	res1, res2 := r.data[r.cursor] & 15, r.data[r.cursor] >> 4
	r.cursor++
	return res1, res2
}

func (r *BytesReader) Read8() uint8 {
	r.cursor++
	return r.data[r.cursor-1]
}

func (r *BytesReader) Readx(x int) []byte {
	r.cursor += x
	return r.data[r.cursor-x:r.cursor]
}

func (r *BytesReader) ReadUTF8() []byte {
	if r.data[r.cursor] < 128 { // length 1
		r.cursor++
		return r.data[r.cursor-1:r.cursor]
	}
	if r.data[r.cursor] & 32 == 0 { // length 2
		r.cursor += 2
		return r.data[r.cursor-2:r.cursor]
	} else {
		r.cursor += 3
		return r.data[r.cursor-3:r.cursor]
	}
}

func (r *BytesReader) Read16() uint16 {
	r.cursor += 2
	return uint16(r.data[r.cursor-2]) | uint16(r.data[r.cursor-1])<<8
}

// If it's less than 255 then it's encoded in the 1st byte, otherwise 1st byte is 255 and it's encoded in two more bytes
// This is only useful if it is expected that the value will be <255 more than half the time
func (r *BytesReader) Read16Variable() uint16 {
	v := r.Read8()
	if v < 255 {
		return uint16(v)
	}
	return r.Read16()
}

func (r *BytesReader) ReadInt16Variable() int16 {
	v := r.Read8()
	if v < 255 {
		return int16(v) - 127
	}
	return int16(r.Read16())
}

func (r *BytesReader) Read24() uint32 {
	r.cursor += 3
	return uint32(r.data[r.cursor-3]) | uint32(r.data[r.cursor-2])<<8 | uint32(r.data[r.cursor-1])<<16
}

func (r *BytesReader) Read32() uint32 {
	r.cursor += 4
	return uint32(r.data[r.cursor-4]) | uint32(r.data[r.cursor-3])<<8 | uint32(r.data[r.cursor-2])<<16 | uint32(r.data[r.cursor-1])<<24
}

func (r *BytesReader) ReadFloat32() float32 {
	return math.Float32frombits(r.Read32())
}

func (r *BytesReader) Read48() uint64 {
	r.cursor += 6
	return uint64(r.data[r.cursor-6]) | uint64(r.data[r.cursor-5])<<8 | uint64(r.data[r.cursor-4])<<16 | uint64(r.data[r.cursor-3])<<24 | uint64(r.data[r.cursor-2])<<32 | uint64(r.data[r.cursor-1])<<40
}

func (r *BytesReader) Read64() uint64 {
	r.cursor += 8
	return uint64(r.data[r.cursor-8]) | uint64(r.data[r.cursor-7])<<8 | uint64(r.data[r.cursor-6])<<16 | uint64(r.data[r.cursor-5])<<24 | uint64(r.data[r.cursor-4])<<32 | uint64(r.data[r.cursor-3])<<40 | uint64(r.data[r.cursor-2])<<48 | uint64(r.data[r.cursor-1])<<56
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (r *BytesReader) Read64Variable() uint64 {
	s1 := int(r.Read8())
	var res1 uint64
	switch s1 {
		case 1: res1 = uint64(r.data[r.cursor])
		case 2: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8
		case 3: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16
		case 4: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24
		case 5: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32
		case 6: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40
		case 7: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40 | uint64(r.data[r.cursor+6])<<48
		case 8: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40 | uint64(r.data[r.cursor+6])<<48 | uint64(r.data[r.cursor+7])<<56
	}
	r.cursor += s1
	return res1
}

// The first byte stores the bit length of the two integers. Then come the two integers. Length is only 1 byte more than the smallest representation of both integers.
func (r *BytesReader) Read64Variable2() (uint64, uint64) {
	s2 := r.Read8()
	s1 := s2 >> 4
	s2 &= 15
	var res1, res2 uint64
	switch s1 {
		case 1: res1 = uint64(r.data[r.cursor])
		case 2: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8
		case 3: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16
		case 4: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24
		case 5: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32
		case 6: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40
		case 7: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40 | uint64(r.data[r.cursor+6])<<48
		case 8: res1 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40 | uint64(r.data[r.cursor+6])<<48 | uint64(r.data[r.cursor+7])<<56
	}
	r.cursor += int(s1)
	switch s2 {
		case 1: res2 = uint64(r.data[r.cursor])
		case 2: res2 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8
		case 3: res2 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16
		case 4: res2 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24
		case 5: res2 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32
		case 6: res2 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40
		case 7: res2 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40 | uint64(r.data[r.cursor+6])<<48
		case 8: res2 = uint64(r.data[r.cursor]) | uint64(r.data[r.cursor+1])<<8 | uint64(r.data[r.cursor+2])<<16 | uint64(r.data[r.cursor+3])<<24 | uint64(r.data[r.cursor+4])<<32 | uint64(r.data[r.cursor+5])<<40 | uint64(r.data[r.cursor+6])<<48 | uint64(r.data[r.cursor+7])<<56
	}
	r.cursor += int(s2)
	return res1, res2
}

func (r *BytesReader) ReadFloat64() float64 {
	return math.Float64frombits(r.Read64())
}

func (r *BytesReader) ReadString8() string {
	return string(r.Readx(int(r.Read8())))
}

func (r *BytesReader) ReadString16() string {
	return string(r.Readx(int(r.Read16())))
}

func (r *BytesReader) ReadString32() string {
	return string(r.Readx(int(r.Read32())))
}

// 12 bits for uint16 and 4 bits for uint8
func (r *BytesReader) Read12() (uint16, uint16) {
	res := uint16(r.data[r.cursor]) | uint16(r.data[r.cursor+1])<<8
	r.cursor += 2
	return res & 4095, res >> 12
}

func (r *BytesReader) ReadSpecial() (uint8, bool, bool, bool, bool) {
	c := r.data[r.cursor]
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
	r.cursor++
	return c & 7, b1, b2, b3, b4
}

func (r *BytesReader) ReadSpecial2() (uint8, uint8, uint8, bool) {
	c := r.data[r.cursor]
	var b1 bool
	if c & 128 > 0 {
		b1 = true
	}
	v1 := c & 7
	v2 := (c >> 3) & 3
	v3 := (c >> 5) & 3
	r.cursor++
	return v1, v2, v3, b1
}
