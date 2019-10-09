// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package protocol

import (
    "unsafe"
)

type RequestHeader struct {
    MagicCode uint16
    Version   uint16
    CRC       int16
    Reserve   int16
    Length    int64
}

type ResponseHeader struct {
    MagicCode uint16
    Version   uint16
    CRC       int16
    Reserve   int16
    Length    int64
}

var (
    RequestHeadSize    = unsafe.Sizeof(RequestHeader{})
    ResponseHeaderSize = unsafe.Sizeof(ResponseHeader{})
)
