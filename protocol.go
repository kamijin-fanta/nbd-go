package nbd

import (
	"fmt"
	"strconv"
)

const (
	nbdMagic             = 0x4e42444d41474943
	optMagic             = 0x49484156454F5054
	repMagic             = 0x0003e889045565a9
	reqMagic             = 0x25609513
	simpleReplyMagic     = 0x67446698
	structuredReplyMagic = 0x668e33ef
	flagFixedNewstyle    = 1 << 0
	flagNoZeroes         = 1 << 1
	flagDefaults         = flagFixedNewstyle | flagNoZeroes
	maxOptionLength      = 4 << 10
)

type cmd uint16

const (
	cmdRead        cmd = 0
	cmdWrite           = 1
	cmdDisc            = 2
	cmdFlush           = 3
	cmdTrim            = 4
	cmdCache           = 5
	cmdWriteZeroes     = 6
	cmdBlockStatus     = 7
	cmdResize          = 8
)

type Errno uint32

const (
	_ Errno = (1 << 31) + iota
	ErrUnsup
	ErrPolicy
	ErrInvalid
	ErrPlatform
	ErrTLSReqd
	ErrUnknown
	ErrShutdown
	ErrBlockSizeReqd
	ErrTooBig
)

func (e Errno) String() string {
	switch e {
	case ErrUnsup:
		return "ERR_UNSUP"
	case ErrPolicy:
		return "ERR_POLICY"
	case ErrInvalid:
		return "ERR_INVALID"
	case ErrPlatform:
		return "ERR_PLATFORM"
	case ErrTLSReqd:
		return "ERR_TLS_REQD"
	case ErrUnknown:
		return "ERR_UNKNOWN"
	case ErrShutdown:
		return "ERR_SHUTDOWN"
	case ErrBlockSizeReqd:
		return "ERR_BLOCK_SIZE_REQD"
	case ErrTooBig:
		return "ERR_TOO_BIG"
	default:
		return "0x" + strconv.FormatUint(uint64(e), 16)
	}
}

func (e Errno) Error() string {
	return fmt.Sprintf("nbd protocol error: %s", e.String())
}

type Errornum uint32

// See https://manpages.debian.org/stretch/manpages-dev/errno.3.en.html for a
// description of error numbers.
const (
	EPERM     Errornum = 1
	EIO       Errornum = 5
	ENOMEM    Errornum = 12
	EINVAL    Errornum = 22
	ENOSPC    Errornum = 28
	EOVERFLOW Errornum = 75
	ESHUTDOWN Errornum = 108
)

var errStr = map[Errornum]string{
	EPERM:     "Operation not permitted",
	EIO:       "Input/output error",
	ENOMEM:    "Cannot allocate memory",
	EINVAL:    "Invalid argument",
	ENOSPC:    "No space left on device",
	EOVERFLOW: "Value too large for defined data type",
	ESHUTDOWN: "Cannot send after transport endpoint shutdown",
}

func (e Errornum) Error() string {
	if msg, ok := errStr[e]; ok {
		return msg
	}
	return fmt.Sprintf("NBD_ERROR(%d)", uint32(e))
}

// Errornum returns e.
func (e Errornum) Errno() Errornum {
	return e
}

type errf struct {
	errno Errornum
	error
}

func (e errf) Errno() Errornum {
	return e.errno
}
