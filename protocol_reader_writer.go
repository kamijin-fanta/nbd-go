package nbd

import (
	"encoding/binary"
	"fmt"
	"io"
)

type ProtocolReaderWriter struct {
	rw          io.ReadWriter
	handleError func(error)
}

func NewProtocolReaderWriter(rw io.ReadWriter, handleError func(error)) ProtocolReaderWriter {
	return ProtocolReaderWriter{
		rw:          rw,
		handleError: handleError,
	}
}

func (rw *ProtocolReaderWriter) Write(buf []byte) {
	_, err := rw.rw.Write(buf)
	if rw.handleError != nil {
		rw.handleError(err)
	}
}

func (rw *ProtocolReaderWriter) WriteString(buf string) {
	rw.Write([]byte(buf))
}

func (rw *ProtocolReaderWriter) Read(buf []byte) {
	_, err := io.ReadFull(rw.rw, buf)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	if rw.handleError != nil {
		rw.handleError(err)
	}
}

func (rw *ProtocolReaderWriter) Discard(n uint32) {
	buf := make([]byte, 512)
	for n > 0 {
		if n > uint32(len(buf)) {
			buf = buf[:n]
		}
		rw.Read(buf)
		n -= uint32(len(buf))
	}
}

func (rw *ProtocolReaderWriter) Uint16() uint16 {
	var b [2]byte
	rw.Read(b[:])
	return binary.BigEndian.Uint16(b[:])
}

func (rw *ProtocolReaderWriter) Uint32() uint32 {
	var b [4]byte
	rw.Read(b[:])
	return binary.BigEndian.Uint32(b[:])
}

func (rw *ProtocolReaderWriter) Uint64() uint64 {
	var b [8]byte
	rw.Read(b[:])
	return binary.BigEndian.Uint64(b[:])
}

func (rw *ProtocolReaderWriter) WriteUint16(v uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], v)
	rw.Write(b[:])
}

func (rw *ProtocolReaderWriter) WriteUint32(v uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	rw.Write(b[:])
}

func (rw *ProtocolReaderWriter) WriteUint64(v uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	rw.Write(b[:])
}

func (rw *ProtocolReaderWriter) ReadOption() (optionType, interface{}, errno) {
	magic := rw.Uint64()
	if magic != optMagic && rw.handleError != nil {
		rw.handleError(fmt.Errorf("invalid option magic 0x%x", magic))
	}
	option := optionType(rw.Uint32())
	length := rw.Uint32()
	if length > maxOptionLength {
		return option, nil, errTooBig
	}
	var o RequestOption
	switch option {
	case cOptInfo:
		o = &optReqInfo{done: false}
	case cOptGo:
		o = &optReqInfo{done: true}
	}
	if o == nil {
		return option, nil, errUnsup
	}
	err := o.ReadFrom(rw, length)
	return option, o, err
}

// respondErr writes an error respons to e, based on handle an err.
func (rw *ProtocolReaderWriter) WriteErrResponse(handle uint64, err error) {
	code := EIO
	if e, ok := err.(Error); ok {
		code = e.Errno()
	}
	rep := simpleReply{
		errno:  uint32(code),
		handle: handle,
		length: 0,
	}
	rep.WriteTo(rw)
}
