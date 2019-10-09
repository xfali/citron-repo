// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package ioutil

import (
    "bytes"
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