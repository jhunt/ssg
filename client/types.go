package client

import "time"

type StreamInfo struct {
	ID      string    `json:"id"`
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

type UploadData struct {
	Data string `json:"data"`
	EOF  bool   `json:"eof"`
}

type Response struct {
	OK string `json:"ok"`
}
