// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package binary

import (
    "bytes"
    "citron-repo/ioutil"
    "citron-repo/protocol"
    "citron-repo/transport"
    "encoding/binary"
    "errors"
    "github.com/xfali/goutils/log"
    "io"
)

const (
    MagicCode       = 0xC100
    Version         = 0x01
    PkgReadBufSize  = 32 * 1024
    PkgWriteBufSize = 32 * 1024
)

type BinaryServer struct {
    magicCode   uint16
    version     uint16
    bodyHandler BodyHandler
    transport   *transport.TcpTransport
    connList    []*binaryConn
}

type binaryConn struct {
    readChan  chan []byte
    writeChan chan []byte
    stopChan  chan bool
}

type PackageWriter func(size int64, reader io.Reader) error
type BodyHandler func(data []byte, resp protocol.Response) error

type Opt func(s *BinaryServer)

func SetMagicCode(magicCode uint16) Opt {
    return func(s *BinaryServer) {
        s.magicCode = magicCode
    }
}

func SetVersion(version uint16) Opt {
    return func(s *BinaryServer) {
        s.version = version
    }
}

func SetBodyHandler(handler BodyHandler) Opt {
    return func(s *BinaryServer) {
        s.bodyHandler = handler
    }
}

func NewBinaryServer(opts ...Opt) *BinaryServer {
    s := BinaryServer{
        magicCode: MagicCode,
        version:   Version,
    }

    tcp := transport.NewTcpTransport(
        transport.SetReadBufSize(PkgReadBufSize),
        transport.SetPort(":20000"),
        transport.SetListenerFactory(s.createListener),
    )
    s.transport = tcp

    for i := range opts {
        opts[i](&s)
    }

    return &s
}

func (s *BinaryServer) ListenAndServe() {
    s.transport.Startup()
}

func (s *BinaryServer) createListener() transport.DataListener {
    c := &binaryConn{
        readChan:  make(chan []byte),
        writeChan: make(chan []byte),
        stopChan:  make(chan bool),
    }
    s.connList = append(s.connList, c)
    go c.process(s.magicCode, s.version, s.bodyHandler)
    return c.getListener
}

func (s *BinaryServer) Close() error {
    for _, v := range s.connList {
        v.Close()
    }
    return s.transport.Close()
}

func (c *binaryConn) Close() error {
    close(c.stopChan)
    return nil
}

func (c *binaryConn) getListener() (chan<- []byte, <-chan []byte) {
    return c.readChan, c.writeChan
}

func (c *binaryConn) process(magicCode, version uint16, handler BodyHandler) {
    pkg := pkgHandler{
        magicCode:   magicCode,
        version:     version,
        ready:       false,
        readBuf:     make([]byte, PkgReadBufSize),
        writeBuf:    make([]byte, PkgWriteBufSize),
        writer:      c.writeChan,
        bodyHandler: handler,
    }

    for {
        select {
        case <-c.stopChan:
            return
        case d := <-c.readChan:
            log.Debug("receive : %s", string(d))
            err := pkg.process(d)
            if err != nil {
                return
            }
            //s.writeChan <- []byte("server reply: " + string(d))
        }
    }
}

type pkgHandler struct {
    magicCode    uint16
    version      uint16
    writer       chan<- []byte
    ready        bool
    headerOffset int
    readBuf      []byte
    writeBuf     []byte
    header       protocol.RequestHeader
    bodyOffset   int64
    bodyHandler  BodyHandler
}

func (pkg *pkgHandler) reset() {
    pkg.ready = false
    pkg.headerOffset = 0
    pkg.bodyOffset = 0
    pkg.header = protocol.RequestHeader{}
}

func (pkg *pkgHandler) toHeader() error {
    err := binary.Read(bytes.NewReader(pkg.readBuf), binary.BigEndian, &pkg.header)
    if err != nil {
        return err
    }

    errC := pkg.checkHeader()
    if errC != nil {
        return errC
    }

    return nil
}

func (pkg *pkgHandler) checkHeader() error {
    if pkg.header.MagicCode != pkg.magicCode {
        return errors.New("Magic code not match ")
    }
    if pkg.header.Version != pkg.version {
        return errors.New("Version not match ")
    }
    return nil
}

func (pkg *pkgHandler) process(data []byte) error {
    if pkg.ready {
        return pkg.processbody(data)
    } else {
        return pkg.processHeader(data)
    }
}

func (pkg *pkgHandler) processHeader(data []byte) error {
    length := int(protocol.RequestHeadSize) - pkg.headerOffset
    dataLen := len(data)
    copyLen := 0
    leftLen := 0
    if dataLen > length {
        copyLen = length
        leftLen = dataLen - length
        pkg.ready = true
    } else if dataLen < length {
        copyLen = dataLen
        leftLen = -1
        pkg.ready = false
    } else {
        copyLen = dataLen
        leftLen = 0
        pkg.ready = true
    }
    copy(pkg.readBuf, data[:copyLen])
    if pkg.ready {
        if err := pkg.toHeader(); err != nil {
            return err
        }
    }

    if leftLen > 0 {
        return pkg.processbody(data[copyLen:])
    }
    return nil
}

func (pkg *pkgHandler) processbody(data []byte) error {
    if !pkg.ready {
        return errors.New("package not ready")
    }
    if pkg.bodyHandler == nil {
        panic("body handler is nil")
    }
    length := int64(len(data))
    left := pkg.header.Length - pkg.bodyOffset
    if length <= left {
        pkg.bodyOffset += length
        err := pkg.bodyHandler(data, pkg.write)
        if err != nil {
            return err
        }
    } else if length > left {
        pkg.bodyOffset += left
        err := pkg.bodyHandler(data[:left], pkg.write)
        if err != nil {
            return err
        }
    }

    //wait for next package
    if pkg.header.Length == pkg.bodyOffset {
        pkg.reset()
    }

    if length > left {
        return pkg.process(data[left:])
    }
    return nil
}

func (pkg *pkgHandler) createHeader(size int64) protocol.ResponseHeader {
    return protocol.ResponseHeader{
        MagicCode: pkg.magicCode,
        Version:   pkg.version,
        Length:    size,
    }
}

func (pkg *pkgHandler) Write(d []byte) (n int, err error) {
    pkg.writer <- d
    return len(d), nil
}

func (pkg *pkgHandler) write(size int64, reader io.Reader) (err error) {
    //write header
    writer := bytes.NewBuffer(pkg.writeBuf)
    header := pkg.createHeader(size)
    err = binary.Write(writer, binary.BigEndian, header)
    if err != nil {
        return err
    }

    _, err = pkg.Write(writer.Bytes())
    if err != nil {
        return err
    }

    //write body
    if reader != nil {
        writer.Reset()

        _, err = ioutil.CopyNWithBuffer(pkg, reader, header.Length, pkg.writeBuf)
        if err != nil {
            return err
        }
    }

    return nil
}

func dummyHandler(data []byte, writer PackageWriter) error {

}
