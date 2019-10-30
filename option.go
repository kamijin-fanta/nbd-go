package nbd

import (
	"bytes"
	"errors"
	"fmt"
)

type optionType uint32

const (
	cOptExportName      optionType = 1
	cOptAbort                      = 2
	cOptList                       = 3
	cOptStartTLS                   = 5
	cOptInfo                       = 6
	cOptGo                         = 7
	cOptStructuredReply            = 8
	cOptListMetaContext            = 9
	cOptSetMetaContext             = 10
)

type optionReplyCode uint32

const (
	cRepAck    optionReplyCode = 1
	cRepServer                 = 2
	cRepInfo                   = 3
)

type OptionInfoCode uint16

const (
	cInfoExport      OptionInfoCode = 0
	cInfoName                       = 1
	cInfoDescription                = 2
	cInfoBlockSize                  = 3
)

/********** RequestOption **********/

type RequestOption interface {
	Code() optionType
	ReadFrom(rw *ProtocolReaderWriter, l uint32) Errno
	WriteTo(rw *ProtocolReaderWriter)
}

type optReqInfo struct {
	done bool
	name string
	reqs []uint16
}

func (o *optReqInfo) Code() optionType {
	if o.done {
		return cOptGo
	}
	return cOptInfo
}

func (o *optReqInfo) ReadFrom(rw *ProtocolReaderWriter, l uint32) Errno {
	nlen := rw.Uint32()
	if l < 6 || nlen > l-6 {
		return ErrInvalid
	}
	name := make([]byte, nlen)
	rw.Read(name)
	o.name = string(name)
	nreqs := rw.Uint16()
	if (l-nlen-6)%2 != 0 || (l-nlen-6)/2 != uint32(nreqs) {
		return ErrInvalid
	}
	for ; nreqs > 0; nreqs-- {
		o.reqs = append(o.reqs, rw.Uint16())
	}
	return 0
}

func (o *optReqInfo) WriteTo(rw *ProtocolReaderWriter) {
	rw.WriteUint32(uint32(len(o.name)))
	rw.WriteString(o.name)
	rw.WriteUint16(uint16(len(o.reqs)))
	for _, r := range o.reqs {
		rw.WriteUint16(r)
	}
}

/********** ReplyOption **********/

type ReplyOption interface {
	Code() optionReplyCode
	ReadFrom(rw *ProtocolReaderWriter, l uint32)
	WriteTo(rw *ProtocolReaderWriter)
}

func OptionReplyWriteTo(rw *ProtocolReaderWriter, option optionType, reply ReplyOption) {
	rw.WriteUint64(repMagic)
	rw.WriteUint32(uint32(option))
	rw.WriteUint32(uint32(reply.Code()))

	var b bytes.Buffer
	var tmpRw = &ProtocolReaderWriter{
		rw:          &b,
		handleError: nil,
	}
	reply.WriteTo(tmpRw)

	by := b.Bytes()
	rw.WriteUint32(uint32(len(by)))
	rw.Write(by)
	fmt.Printf("OptionReplyWriteTo: type=%d code=%d len=%d buff=%v\n", option, reply.Code(), len(by), by)
}

type repAck struct{}

func (r *repAck) Code() optionReplyCode { return cRepAck }

func (r *repAck) WriteTo(rw *ProtocolReaderWriter) {}

func (r *repAck) ReadFrom(rw *ProtocolReaderWriter, l uint32) {
	if l != 0 {
		rw.handleError(errors.New("invalid ack response"))
	}
}

type infoExport struct {
	size  uint64
	flags uint16
}

func (r *infoExport) Code() optionReplyCode { return cRepInfo }

func (r *infoExport) WriteTo(rw *ProtocolReaderWriter) {
	rw.WriteUint16(uint16(cInfoExport))
	rw.WriteUint64(r.size)
	rw.WriteUint16(r.flags)
}

func (r *infoExport) ReadFrom(rw *ProtocolReaderWriter, l uint32) {
	if l != 10 {
		rw.handleError(errors.New("invalid length for info reply"))
	}
	r.size = rw.Uint64()
	r.flags = rw.Uint16()
}

type infoName struct {
	name string
}

func (r *infoName) Code() optionReplyCode { return cRepInfo }

func (r *infoName) WriteTo(rw *ProtocolReaderWriter) {
	rw.WriteUint16(cInfoName)
	rw.WriteString(r.name)
}

func (r *infoName) ReadFrom(rw *ProtocolReaderWriter, l uint32) {
	if l > (4 << 10) {
		rw.handleError(errors.New("name too large"))
	}
	b := make([]byte, l)
	rw.Read(b)
	r.name = string(b)
}

type infoDescription struct {
	description string
}

func (r *infoDescription) Code() optionReplyCode { return cRepInfo }

func (r *infoDescription) WriteTo(rw *ProtocolReaderWriter) {
	rw.WriteUint16(cInfoDescription)
	rw.WriteString(r.description)
}

func (r *infoDescription) ReadFrom(rw *ProtocolReaderWriter, l uint32) {
	if l > (4 << 10) {
		rw.handleError(errors.New("description too large"))
	}
	b := make([]byte, l)
	rw.Read(b)
	r.description = string(b)
}

type infoBlockSize struct {
	min       uint32
	preferred uint32
	max       uint32
}

func (r *infoBlockSize) Code() optionReplyCode { return cRepInfo }

func (r *infoBlockSize) WriteTo(rw *ProtocolReaderWriter) {
	rw.WriteUint16(cInfoBlockSize)
	rw.WriteUint32(r.min)
	rw.WriteUint32(r.preferred)
	rw.WriteUint32(r.max)
}

func (r *infoBlockSize) ReadFrom(rw *ProtocolReaderWriter, l uint32) {
	if l != 12 {
		rw.handleError(errors.New("invalid length for block size info"))
	}
	r.min = rw.Uint32()
	r.preferred = rw.Uint32()
	r.max = rw.Uint32()
}

type repError struct{
	errorCode Errno
}

func (r *repError) Code() optionReplyCode { return optionReplyCode(r.errorCode) }

func (r *repError) WriteTo(rw *ProtocolReaderWriter) {}

func (r *repError) ReadFrom(rw *ProtocolReaderWriter, l uint32) {}
