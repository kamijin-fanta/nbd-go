package nbd

import "net"

type DeviceConnectionFactory interface {
	NewClient(remoteAddr net.Addr) DeviceConnection
}
type DeviceConnection interface {
	ExportList() ([]string, Errno)
	Info(export string) (name, description string, totalSize uint64, blockSize uint32, err Errno)
	Read(offset uint64, length uint32) ([]byte, Errno)
	Write(offset uint64, buff []byte) Errno
	Flush() Errno
}
