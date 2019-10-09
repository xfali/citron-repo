// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package protocol

type Command func(data []byte, writer chan<- []byte) error

const (
    DebugCommandID = iota
)

var cmdMap = map[int16]Command{}

func FindCommand(cmd int16) Command {
    //return cmdMap[cmd]
    return DebugCommand
}

func DebugCommand(data []byte, writer chan<- []byte) error {
    writer <- []byte("debug: " + string(data))
    return nil
}
