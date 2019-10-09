// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package ioutil

import (
    "io"
    "math"
)

func Copy(writer io.Writer, reader io.Reader) (written int64, err error) {
    size := 32 * 1024
    buf := make([]byte, size)
    return CopyWithBuffer(writer, reader, buf)
}

func CopyWithBuffer(writer io.Writer, reader io.Reader, buf []byte) (written int64, err error) {
    //for {
    //    nr, er := reader.Read(buf)
    //    if nr > 0 {
    //        nw, ew := writer.Write(buf[0:nr])
    //        if nw > 0 {
    //            written += int64(nw)
    //        }
    //        if ew != nil {
    //            err = ew
    //            break
    //        }
    //        if nr != nw {
    //            err = io.ErrShortWrite
    //            break
    //        }
    //    }
    //    if er != nil {
    //        if er != io.EOF {
    //            err = er
    //        }
    //        break
    //    }
    //}
    //return written, err
    return CopyNWithBuffer(writer, reader, math.MaxInt64, buf)
}

func CopyN(writer io.Writer, reader io.Reader, n int64) (written int64, err error) {
    size := 32 * 1024
    buf := make([]byte, size)
    return CopyNWithBuffer(writer, reader, n, buf)
}

func CopyNWithBuffer(writer io.Writer, reader io.Reader, n int64, buf []byte) (written int64, err error) {
    if n <= 0 {
        return 0, nil
    }
    size := len(buf)
    reads := size
    for {
        if n-written < int64(size) {
            reads = int(n - written)
        }
        nr, er := reader.Read(buf[0:reads])
        if nr > 0 {
            nw, ew := writer.Write(buf[0:nr])
            if nw > 0 {
                written += int64(nw)
            }
            if ew != nil {
                err = ew
                break
            }
            if nr != nw {
                err = io.ErrShortWrite
                break
            }
        }
        if er != nil {
            if er != io.EOF {
                err = er
            }
            break
        }
        if written == n {
            break
        }
    }
    return written, err
}
