// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package binary

import "io"

type Request struct {
    Body   io.Reader
}

type Response struct {
    Length int64
    Body   io.Reader
}
