package api

import "github.com/jhunt/shield-storage-gateway/vault"

type UploadRequest struct {
	Prefix string `json:"prefix"`
	Agent  string `json:"agent"`
}

type DownloadRequest struct {
	Path  string `json:"path"`
	Agent string `json:"agent"`
}

type UploadChunk struct {
	Sequence uint   `json:"seq"`
	Size     uint   `json:"size"`
	Data     string `json:"data"`
	EOF      bool   `json:"eof"`
}

type StreamKey struct {
	ID       string `json:"id"`
	Token    string `json:"token"`
	Lifetime uint   `json:"lifetime"`
}

type StreamConfig struct {
	Compression string `json:"compression"`
	Encryption  string `json:"encryption"`
	VaultClient *vault.Client
}
