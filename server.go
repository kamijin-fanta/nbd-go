package nbd

import (
	"fmt"
	"io"
	"net"
	"sync"
)

func ListenAndServe(listen net.Listener, factory DeviceConnectionFactory) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}
		wg.Add(1)
		deviceConnection := factory.NewClient(conn.RemoteAddr())
		go func() {
			defer func() {
				deviceConnection.Close()
				conn.Close()
				wg.Done()
			}()
			err := Handshake(conn, deviceConnection)
			if err != nil {
				fmt.Printf("error on handshake %v\n", err)
				return
			}
			err = Serve(conn, deviceConnection)
			if err != nil {
				fmt.Printf("error on transmission %v\n", err)
				return
			}
		}()
	}
}

type ConnectionState struct {
	buff []byte
}

func Handshake(conn net.Conn, deviceConnection DeviceConnection) error {
	return ConnectionWrapper(conn, func(e *ProtocolReaderWriter) {
		e.WriteUint64(nbdMagic)
		e.WriteUint64(optMagic)
		e.WriteUint16(flagDefaults)

		clientFlags := e.Uint32()

		if clientFlags & ^uint32(flagDefaults) != 0 {
			e.handleError(fmt.Errorf("handshake aborted due to unknown handshake flags 0x%d", clientFlags & ^uint32(flagDefaults)))
		}
		if clientFlags != flagDefaults {
			e.handleError(fmt.Errorf("refusing deprecated handshake flags 0x%x", clientFlags))
		}

		for {
			code, o, err := e.ReadOption()
			if err != 0 {
				OptionReplyWriteTo(e, code, &repError{err})
				continue
			}
			switch o := o.(type) {
			// todo more options
			case *optReqInfo:
				name, description, totalSize, blockSize, errno := deviceConnection.Info(o.name)
				if errno != 0 {
					OptionReplyWriteTo(e, code, &repError{errno})
					continue
				}

				OptionReplyWriteTo(e, code, &infoExport{
					totalSize,
					0,
				})

				for _, r := range o.reqs {
					switch OptionInfoCode(r) {
					case cInfoExport:
						// already send option
					case cInfoName:
						OptionReplyWriteTo(e, code, &infoName{name})
					case cInfoDescription:
						OptionReplyWriteTo(e, code, &infoDescription{description})
					case cInfoBlockSize:
						OptionReplyWriteTo(e, code, &infoBlockSize{
							blockSize,
							blockSize,
							blockSize,
						})
					}
				}
				OptionReplyWriteTo(e, code, &repAck{})
				if o.done {
					return
				}
			}
		}
	})
}
func Serve(conn net.Conn, deviceConnection DeviceConnection) error {
	return ConnectionWrapper(conn, func(rw *ProtocolReaderWriter) {
		var req request
		for {
			if err := req.ReadFrom(rw); err != nil {
				rw.WriteErrResponse(req.handle, err)
				continue
			}
			switch req.typ {
			case cmdRead:
				if req.length == 0 {
					rw.WriteErrResponse(req.handle, EINVAL)
					continue
				}
				buf, err := deviceConnection.Read(req.offset, req.length)
				if err != 0 {
					rw.WriteErrResponse(req.handle, err)
					continue
				}
				(&simpleReply{0, req.handle, buf, 0}).WriteTo(rw)
			case cmdWrite:
				if req.length == 0 {
					rw.WriteErrResponse(req.handle, EINVAL)
					continue
				}
				err := deviceConnection.Write(req.offset, req.data)
				if err != 0 {
					rw.WriteErrResponse(req.handle, err)
					continue
				}
				(&simpleReply{0, req.handle, nil, 0}).WriteTo(rw)
			case cmdDisc:
				return
			case cmdFlush:
				if req.length != 0 || req.offset != 0 {
					rw.WriteErrResponse(req.handle, EINVAL)
					continue
				}
				err := deviceConnection.Flush()
				if err != 0 {
					rw.WriteErrResponse(req.handle, err)
					continue
				}
				(&simpleReply{0, req.handle, nil, 0}).WriteTo(rw)
			default:
				rw.WriteErrResponse(req.handle, EINVAL)
			}
		}
	})
}

func ConnectionWrapper(rw io.ReadWriter, f func(c *ProtocolReaderWriter)) (err error) {
	sentinel := new(uint8)
	defer func() {
		if v := recover(); err == nil && v != sentinel {
			if e, ok := v.(error); ok {
				err = e
			}
		}
	}()
	errorHandle := func(e error) {
		if e != nil {
			err = e
			panic(sentinel)
		}
	}
	prw := NewProtocolReaderWriter(rw, errorHandle)
	f(&prw)
	return
}
