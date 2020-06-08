package ssg

import (
	"sync"
	"time"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/vault"
)

type stream struct {
	id    string
	canon string

	secret  string
	leased  time.Time
	expires time.Time
	renewal time.Duration

	segments int
	writer provider.Uploader
	reader provider.Downloader
}

type bucket struct {
	key         string
	name        string
	description string

	compression string
	encryption  string

	provider provider.Provider
	vault    vault.Vault
}

type Server struct {
	Cluster       string
	Bind          string
	ControlTokens []string
	MonitorTokens []string
	MaxLease      time.Duration
	SweepInterval time.Duration

	lock      sync.Mutex
	buckets   []*bucket
	uploads   map[string]*stream
	downloads map[string]*stream
}

func (s Server) bucket(key string) *bucket {
	for i := range s.buckets {
		if s.buckets[i].key == key {
			return s.buckets[i]
		}
	}
	return nil
}
