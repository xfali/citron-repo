// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package binary

import "io"

type Request interface {
    Read(w io.Writer) error
}

type Response struct {
    Length int64
    Write  func(w io.Reader) error
}
