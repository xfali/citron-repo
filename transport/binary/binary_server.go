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
    "strings"
)

const (
    MagicCode       = 0xC100
    Version         = 0x01
    PkgReadBufSize  = 32 * 1024
    PkgWriteBufSize = 32 * 1024
)

type BinaryServer struct {
    port           string
    magicCode      uint16
    version        uint16
    requestHandler RequestHandler
    transport      *transport.TcpTransport
    connList       []*binaryConn
}

type binaryConn struct {
    readChan  chan []byte
    writeChan chan []byte
    stopChan  chan bool
}

type PackageWriter func(size int64, reader io.Reader) error

type RequestHandler interface {
    io.Writer
    OnePackage(PackageWriter) error
}

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

func SetRequestHandler(handler RequestHandler) Opt {
    return func(s *BinaryServer) {
        s.requestHandler = handler
    }
}

func SetPort(port string) Opt {
    return func(s *BinaryServer) {
        s.port = port
    }
}

func NewBinaryServer(opts ...Opt) *BinaryServer {
    s := BinaryServer{
        magicCode: MagicCode,
        version:   Version,
    }

    for i := range opts {
        opts[i](&s)
    }
    if s.port == "" {
        s.port = ":20000"
    }

    if s.requestHandler == nil {
        d := DummyHandler("dummy")
        s.requestHandler = &d
    }

    tcp := transport.NewTcpTransport(
        transport.SetReadBufSize(PkgReadBufSize),
        transport.SetPort(s.port),
        transport.SetListenerFactory(s.createListener),
    )
    s.transport = tcp

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
    go c.process(s.magicCode, s.version, s.requestHandler)
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

func (c *binaryConn) getListener() (chan<- []byte, <-chan []byte, <-chan bool) {
    return c.readChan, c.writeChan, c.stopChan
}

func (c *binaryConn) process(magicCode, version uint16, handler RequestHandler) {
    pkg := pkgHandler{
        magicCode:      magicCode,
        version:        version,
        ready:          false,
        readBuf:        make([]byte, PkgReadBufSize),
        writeBuf:       make([]byte, PkgWriteBufSize),
        reader:         c.readChan,
        writer:         c.writeChan,
        requestHandler: handler,
    }

    for {
        select {
        case <-c.stopChan:
            return
        case d := <-c.readChan:
            log.Debug("receive : %s", string(d))
            err := pkg.next(d)
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
    reader         <-chan []byte
    writer         chan<- []byte
    ready          bool
    headerOffset   int
    readBuf        []byte
    writeBuf       []byte
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
    err := binary.Read(bytes.NewReader(pkg.readBuf), binary.BigEndian, &pkg.header)
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
    copy(pkg.readBuf[pkg.headerOffset:], data[:copyLen])
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
        pkg.requestHandler.OnePackage(pkg.write)
        //prepare for next package
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
    pkg.writer <- d
    return len(d), nil
}

func (pkg *pkgHandler) write(size int64, reader io.Reader) (err error) {
    //write header
    writer := bytes.Buffer{}
    header := pkg.createHeader(size)
    err = binary.Write(&writer, binary.BigEndian, header)
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

type DummyHandler string

func (d *DummyHandler) Write(p []byte) (n int, err error) {
    *d = DummyHandler(p)
    return len(p), nil
}

func (d *DummyHandler) OnePackage(w PackageWriter) error {
    s := string(*d)
    return w(int64(len(s)), strings.NewReader(s))
}
