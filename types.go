package nbd

import "errors"

type Error interface {
	Error() string
	Errno() Errno
}

type request struct {
	flags  uint16
	typ    cmd
	handle uint64
	offset uint64
	length uint32
	data   []byte
}

func (r *request) WriteTo(rw *ProtocolReaderWriter) {
	rw.WriteUint32(reqMagic)
	rw.WriteUint16(r.flags)
	rw.WriteUint16(uint16(r.typ))
	rw.WriteUint64(r.handle)
	rw.WriteUint64(r.offset)
	rw.WriteUint32(uint32(len(r.data)))
	rw.Write(r.data)
}

func (r *request) ReadFrom(rw *ProtocolReaderWriter) Error {
	if rw.Uint32() != reqMagic {
		rw.handleError(errors.New("invalid magic for request"))
	}
	r.flags = rw.Uint16()
	r.typ = cmd(rw.Uint16())
	r.handle = rw.Uint64()
	r.offset = rw.Uint64()
	r.length = rw.Uint32()
	if r.offset&(1<<63) != 0 {
		return EOVERFLOW
	}
	if r.typ != cmdWrite {
		return nil
	}
	if r.length > 4<<20 {
		rw.Discard(r.length)
		return EOVERFLOW
	}
	buf := make([]byte, r.length)
	rw.Read(buf)
	r.data = buf
	return nil
}

type simpleReply struct {
	errno  uint32
	handle uint64
	data   []byte

	length uint32
}

func (r *simpleReply) WriteTo(rw *ProtocolReaderWriter) {
	rw.WriteUint32(simpleReplyMagic)
	rw.WriteUint32(r.errno)
	rw.WriteUint64(r.handle)
	rw.Write(r.data)
}

func (r *simpleReply) ReadFrom(rw *ProtocolReaderWriter) Error {
	if rw.Uint32() != simpleReplyMagic {
		rw.handleError(errors.New("invalid magic for reply"))
	}
	r.handle = rw.Uint64()
	buf := make([]byte, r.length)
	rw.Read(buf)
	return nil
}
