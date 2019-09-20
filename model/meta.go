// Copyright (C) 2019, Xiongfa Li.
// All right reserved.
// @author xiongfa.li
// @version V1.0
// Description: 

package model

import "time"

type FileInfo struct {
    FileName string `json:"filename"`
    FilePath string `json:"filepath"`
    Parent   string `json:"parent"`

    From   string `json:"from"`
    To     string `json:"to"`
    Hidden bool `json:"hidden,omitempty"`

    State int `json:"state"`

    IsDir bool  `json:"isDir"`
    Size  int64 `json:"size"`

    ModTime time.Time `json:"modTime"`

    Checksum     string `json:"checksum,omitempty"`
    ChecksumType string `json:"checksumType,omitempty"`
}