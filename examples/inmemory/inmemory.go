package main

import (
	"context"
	"fmt"
	"github.com/kamijin-fanta/nbd-go"
	"net"
)

func main() {
	network, addr := "tcp", ":8888"
	fmt.Printf("listen on %s %s\n", network, addr)

	var factory nbd.DeviceConnectionFactory = &MemoryDeviceFactory{}

	err := nbd.ListenAndServe(context.Background(), network, addr, factory)
	if err != nil {
		panic(err)
	}
}

type MemoryDeviceFactory struct {
}

func (m *MemoryDeviceFactory) NewClient(remoteAddr net.Addr) nbd.DeviceConnection {
	size := uint64(1024 * 1024 * 500) // 500MB
	return &MemoryDeviceConnection{
		size: size,
		buff: make([]byte, size),
	}
}

type MemoryDeviceConnection struct {
	size uint64
	buff []byte
}

func (m *MemoryDeviceConnection) ExportList() ([]string, nbd.Errno) {
	panic("implement me")
}

func (m *MemoryDeviceConnection) Info(export string) (name, description string, totalSize uint64, blockSize uint32, errno nbd.Errno) {
	return "default", "default exports", m.size, 4096, 0 // 4K Block
}

func (m *MemoryDeviceConnection) Read(offset uint64, length uint32) ([]byte, nbd.Errornum) {
	return m.buff[offset : offset+uint64(length)], 0
}

func (m *MemoryDeviceConnection) Write(offset uint64, buff []byte) nbd.Errornum {
	target := m.buff[offset : offset+uint64(len(buff))]
	copy(target, buff)
	return 0
}

func (m *MemoryDeviceConnection) Flush() nbd.Errornum {
	// nop
	return 0
}

func (m *MemoryDeviceConnection) Close() {
	m.buff = nil
}
