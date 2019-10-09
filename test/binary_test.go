// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package test

import (
    "citron-repo/transport/binary"
    "testing"
)

func TestBinary(t *testing.T) {
    binary.NewBinaryServer(
        binary.SetBodyHandler()
    )
}
