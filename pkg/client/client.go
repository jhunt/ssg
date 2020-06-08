package client

import "time"

type Stream struct {
	Kind    string    `json:"kind"`
	ID      string    `json:"id"`
	Canon   string    `json:"canon"`
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

type Blob struct {
	Segments     int   `json:"segments"`
	Compressed   int64 `json:"compressed"`
	Uncompressed int64 `json:"uncompressed"`
}
