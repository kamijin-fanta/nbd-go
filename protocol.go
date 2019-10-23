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

type errno uint32

const (
	_ errno = (1 << 31) + iota
	errUnsup
	errPolicy
	errInvalid
	errPlatform
	errTLSReqd
	errUnknown
	errShutdown
	errBlockSizeReqd
	errTooBig
)

func (e errno) String() string {
	switch e {
	case errUnsup:
		return "ERR_UNSUP"
	case errPolicy:
		return "ERR_POLICY"
	case errInvalid:
		return "ERR_INVALID"
	case errPlatform:
		return "ERR_PLATFORM"
	case errTLSReqd:
		return "ERR_TLS_REQD"
	case errUnknown:
		return "ERR_UNKNOWN"
	case errShutdown:
		return "ERR_SHUTDOWN"
	case errBlockSizeReqd:
		return "ERR_BLOCK_SIZE_REQD"
	case errTooBig:
		return "ERR_TOO_BIG"
	default:
		return "0x" + strconv.FormatUint(uint64(e), 16)
	}
}

type Errno uint32

// See https://manpages.debian.org/stretch/manpages-dev/errno.3.en.html for a
// description of error numbers.
const (
	EPERM     Errno = 1
	EIO       Errno = 5
	ENOMEM    Errno = 12
	EINVAL    Errno = 22
	ENOSPC    Errno = 28
	EOVERFLOW Errno = 75
	ESHUTDOWN Errno = 108
)

var errStr = map[Errno]string{
	EPERM:     "Operation not permitted",
	EIO:       "Input/output error",
	ENOMEM:    "Cannot allocate memory",
	EINVAL:    "Invalid argument",
	ENOSPC:    "No space left on device",
	EOVERFLOW: "Value too large for defined data type",
	ESHUTDOWN: "Cannot send after transport endpoint shutdown",
}

func (e Errno) Error() string {
	if msg, ok := errStr[e]; ok {
		return msg
	}
	return fmt.Sprintf("NBD_ERROR(%d)", uint32(e))
}

// Errno returns e.
func (e Errno) Errno() Errno {
	return e
}

type errf struct {
	errno Errno
	error
}

func (e errf) Errno() Errno {
	return e.errno
}
