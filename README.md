## Custom

This package makes it easy to save any data to file and read it back quickly and easily, taking advantage of fast binary encoding and Zlib compression. It is much more efficient that using gob encoding and decoding.

The purpose of this package is to enable extremely efficient reading and writing of structures that consist of mostly number types, which let's face it is everything since strings are also based on slices of bytes which themselves are really uint8.

### Writing

Attach to io.writer with default Zlib compression level:

      w := custom.NewWriter(writer)
       
Attach to io.writer with custom Zlib compression level:

      level := 2
      w := custom.NewWriterLevel(writer, level)
       
Write slice of bytes:

     w.Write([]byte("example"))
      
Write boolean value (encoding into 1 byte):

     boolean := true
     w.WriteBool(boolean)
      
Write two boolean values (encoding into 1 byte):

     boolean1, boolean2 := true, false
     w.WriteBool2(boolean1, boolean2)
      
Write two uint8 values in one byte where both are only using 4 bits (0-127):

     fourbit1, fourbit2 := uint8(15), uint8(100)
     w.Write4(fourbit1, fourbit2)
      
Write uint8 value:

     v := uint8(100)
     w.Write8(v)
      
Write uint16 value:

     v := uint16(1000)
      w.Write16(v)
      
Write a variable uint16 value where it is assumed that the value will usually fit into a uint8, but sometimes requires uint16. This will use either 1 byte to encode the uint16 if it is less than 255, otherwise 3 bytes:

     v := uint16(100)
     w.Write16Variable(v)
     
Write a variable int16 value where it is assumed that the value is usually between -127 and 127:

     v := int16(-100)
     w.WriteInt16Variable(v)
     
Write a uint32 value in which only 24 bits need to be encoded:

     v := uint32(20000)
     w.Write24(v)
   
Write a uint32 value:

     v := uint32(100000000)
     w.Write32(v)
     
Write a uint64 value in which only 48 bits need to be encoded:

     v := uint32(10000000000)
     w.Write48(v)
     
Write a uint64 value:

     v := uint64(1000000000000000)
     w.Write64(v)

Write a variable uint64, this will always use 1 byte more than the minimum number of bytes that can be used to encode the integer, i.e. between 2 - 9 bytes depending on the integer.

     v := uint64(1000)
     w.Write64Variable(v)
     
Write two variable uint64s, this will always use 1 byte more than the minimum number of bytes that can be used to encode both integers:

     v1, v2 := uint64(1000), uint64(1)
     w.Write64Variable2(v1, v2)
     
Write a float32:

     w.WriteFloat32(1.234)
     
Write a float64:

     w.WriteFloat64(1.234)
     
Write string with maximum length 255:

     w.WriteString8(`example`)

Write string with maximum length 65535:

     w.WriteString16(`example`)
     
Write string with maximum length 4294967295:

     w.WriteString32(`example`)
     
Write a 12 bit integer and then a 4 bit integer:

     v1, v2 := uint16(380), uint16(7)
     w.Write12(v1, v2)
	 
### Reading

Basically the same but the other way around, e.g.
     
	 bufferSize := 1024 * 20
	 r := custom.NewReader(reader, bufferSize)
	 val := r.Read64Variable()
	 str := r.ReadString16()
	 if err := r.EOF(); err != nil {
		panic(`This should be the end but it's not! Oh no!`)
	 }
	 r.Close()

     
