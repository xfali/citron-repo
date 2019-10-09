// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package ioutil

import (
    "bytes"
    "errors"
    "io"
    "runtime"
    "strconv"
)

func GetGoroutineID() uint64 {
    b := make([]byte, 64)
    runtime.Stack(b, false)
    b = bytes.TrimPrefix(b, []byte("goroutine "))
    b = b[:bytes.IndexByte(b, ' ')]
    n, _ := strconv.ParseUint(string(b), 10, 64)
    return n
}

type ByteWrapper struct {
    B []byte
    L int
    R int
}

func (b *ByteWrapper) Reset() {
    b.L = 0
    b.R = 0
}

func (b *ByteWrapper) Write(p []byte) (n int, err error) {
    if cap(b.B) < b.L + len(p) {
        return 0, errors.New("Out of range ")
    }
    n = copy(b.B[b.L:], p)
    b.L += n
    return
}

func (b *ByteWrapper) Read(p []byte) (n int, err error) {
    if b.L == 0 || b.R == b.L {
        if len(p) == 0 {
            return 0, nil
        }
        return 0, io.EOF
    }
    n = copy(p, b.B[b.R:b.L])
    b.R += n
    return
}

func (b *ByteWrapper) Bytes() []byte {
    return b.B[0:b.L]
}
