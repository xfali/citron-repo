// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package transport

import (
    "bytes"
    "citron-repo/ioutil"
    "citron-repo/protocol"
    "encoding/binary"
    "errors"
    "fmt"
    "github.com/xfali/goutils/log"
    "io"
    "strings"
    "sync"
)

const (
    MagicCode       = 0xC100
    Version         = 0x01
    PkgReadBufSize  = 32 * 1024
    PkgWriteBufSize = 32 * 1024
)

type BinaryServer struct {
    magicCode      uint16
    version        uint16
    requestHandler RequestHandler
    transport      *TcpTransport
    connList       []*binaryConn
}

type binaryConn struct {
    readChan  chan []byte
    writeChan chan []byte
    stopChan  chan bool

    readBufPool  sync.Pool
    writeBufPool sync.Pool
}

type PackageWriter func(size int64, reader io.Reader) error

type RequestHandler interface {
    io.Writer
    Reset()
    OnePackage(PackageWriter) error
}

type BinOpt func(s *BinaryServer)

func SetMagicCode(magicCode uint16) BinOpt {
    return func(s *BinaryServer) {
        s.magicCode = magicCode
    }
}

func SetVersion(version uint16) BinOpt {
    return func(s *BinaryServer) {
        s.version = version
    }
}

func SetRequestHandler(handler RequestHandler) BinOpt {
    return func(s *BinaryServer) {
        s.requestHandler = handler
    }
}

func SetTransport(t *TcpTransport) BinOpt {
    return func(s *BinaryServer) {
        s.transport = t
    }
}

func NewBinaryServer(opts ...BinOpt) *BinaryServer {
    s := BinaryServer{
        magicCode: MagicCode,
        version:   Version,
    }

    for i := range opts {
        opts[i](&s)
    }

    if s.requestHandler == nil {
        d := DummyHandler("")
        s.requestHandler = &d
    }

    if s.transport == nil {
        tcp := NewTcpTransport(
            SetPort(":20001"),
            SetListenerFactory(s.createListener),
        )
        s.transport = tcp
    } else {
        s.transport.connConf.factory = s.createListener
    }

    return &s
}

func (s *BinaryServer) ListenAndServe() {
    s.transport.Startup()
}

func (s *BinaryServer) createListener() Processor {
    c := &binaryConn{
        readChan:  make(chan []byte),
        writeChan: make(chan []byte),
        stopChan:  make(chan bool),

        readBufPool: sync.Pool{New: func() interface{} {
            return make([]byte, PkgReadBufSize)
        }},
        writeBufPool: sync.Pool{New: func() interface{} {
            return make([]byte, PkgWriteBufSize)
        }},
    }
    s.connList = append(s.connList, c)
    go c.process(s.magicCode, s.version, s.requestHandler)
    return c
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

func (c *binaryConn) ReadChan() chan<- []byte {
    return c.readChan
}

func (c *binaryConn) WriteChan() <-chan []byte {
    return c.writeChan
}

func (c *binaryConn) CloseChan() <-chan bool {
    return c.stopChan
}

func (c *binaryConn) AcquireReadBuf() []byte {
    return c.readBufPool.Get().([]byte)[0:PkgReadBufSize]
}

func (c *binaryConn) ReleaseReadBuf(d []byte) {
    c.readBufPool.Put(d)
}

func (c *binaryConn) AcquireWriteBuf() []byte {
    return c.writeBufPool.Get().([]byte)[0:PkgWriteBufSize]
}

func (c *binaryConn) ReleaseWriteBuf(d []byte) {
    c.writeBufPool.Put(d)
}

func (c *binaryConn) process(magicCode, version uint16, handler RequestHandler) {
    pkg := pkgHandler{
        magicCode:      magicCode,
        version:        version,
        ready:          false,
        headerBuf:      make([]byte, PkgReadBufSize),
        conn:           c,
        requestHandler: handler,
    }

    for {
        select {
        case <-c.stopChan:
            return
        case d := <-c.readChan:
            log.Debug("receive : %s", string(d))
            err := pkg.next(d)
            c.ReleaseReadBuf(d)
            if err != nil {
                close(c.stopChan)
                return
            }
            //s.writeChan <- []byte("server reply: " + string(d))
        }
    }
}

type pkgHandler struct {
    magicCode      uint16
    version        uint16
    conn           *binaryConn
    ready          bool
    headerOffset   int
    headerBuf      []byte
    header         protocol.RequestHeader
    bodyOffset     int64
    requestHandler RequestHandler
}

func (pkg *pkgHandler) reset() {
    pkg.ready = false
    pkg.headerOffset = 0
    pkg.bodyOffset = 0
    pkg.header = protocol.RequestHeader{}
}

func (pkg *pkgHandler) toHeader() error {
    err := binary.Read(bytes.NewReader(pkg.headerBuf), binary.BigEndian, &pkg.header)
    if err != nil {
        return err
    }

    errC := pkg.checkHeader()
    if errC != nil {
        return errC
    }

    log.Debug("header is %v", pkg.header)
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

func (pkg *pkgHandler) next(data []byte) error {
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
    if dataLen > length {
        copyLen = length
        pkg.ready = true
    } else if dataLen < length {
        copyLen = dataLen
        pkg.ready = false
    } else {
        copyLen = dataLen
        pkg.ready = true
    }
    copy(pkg.headerBuf[pkg.headerOffset:], data[:copyLen])
    if pkg.ready {
        if err := pkg.toHeader(); err != nil {
            return err
        }
    }
    pkg.headerOffset += copyLen

    if copyLen < dataLen {
        return pkg.processbody(data[copyLen:])
    }
    return nil
}

func (pkg *pkgHandler) processbody(data []byte) error {
    if !pkg.ready {
        return errors.New("package not ready")
    }
    if pkg.requestHandler == nil {
        panic("body handler is nil")
    }

    length := int64(len(data))
    left := pkg.header.Length - pkg.bodyOffset
    if length <= left {
        pkg.bodyOffset += length
        _, err := pkg.requestHandler.Write(data)
        if err != nil {
            return err
        }
    } else if length > left {
        pkg.bodyOffset += left
        _, err := pkg.requestHandler.Write(data[:left])
        if err != nil {
            return err
        }
    }

    if pkg.header.Length == pkg.bodyOffset {
        err := pkg.requestHandler.OnePackage(pkg.write)
        if err != nil {
            return err
        }
        //prepare for next package
        pkg.requestHandler.Reset()
        pkg.reset()
    }

    if length > left {
        return pkg.next(data[left:])
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
    pkg.conn.writeChan <- d
    return len(d), nil
}

func (pkg *pkgHandler) write(size int64, reader io.Reader) (err error) {
    buf := pkg.conn.AcquireWriteBuf()
    //write header
    writer := ioutil.ByteWrapper{B: buf}
    header := pkg.createHeader(size)
    err = binary.Write(&writer, binary.BigEndian, header)
    if err != nil {
        return err
    }
    b := writer.Bytes()
    if len(b) != int(protocol.ResponseHeaderSize) {
        return fmt.Errorf("Response header size error %v ", b)
    }
    _, err = pkg.Write(b)
    if err != nil {
        return err
    }

    //write body
    if reader != nil {
        buf := pkg.conn.AcquireWriteBuf()
        _, err = ioutil.CopyNWithBuffer(pkg, reader, header.Length, buf)
        if err != nil {
            return err
        }
    }

    return nil
}

type DummyHandler string

func (d *DummyHandler) Write(p []byte) (n int, err error) {
    *d = DummyHandler(string(*d) + string(p))
    return len(p), nil
}

func (d *DummyHandler) Reset() {
    *d = ""
}

func (d *DummyHandler) OnePackage(w PackageWriter) error {
    s := string(*d)
    return w(int64(len(s)), strings.NewReader(s))
}
