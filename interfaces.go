package nbd

import "net"

type DeviceConnectionFactory interface {
	NewClient(remoteAddr net.Addr) DeviceConnection
}
type DeviceConnection interface {
	ExportList() ([]string, error)
	Info(export string) (name, description string, totalSize uint64, blockSize uint32)
	Read(offset uint64, length uint32) ([]byte, error)
	Write(offset uint64, buff []byte) error
	Flush() error
}
