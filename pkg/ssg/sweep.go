package ssg

import (
	"time"

	"github.com/jhunt/go-log"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/provider"
)

func (s *Server) Sweep() {
	t := time.NewTicker(s.SweepInterval)
	for range t.C {
		total := 0
		logged := false
		cancel := make(map[string]provider.Uploader)

		s.lock.Lock()
		for id, upload := range s.uploads {
			total += 1
			if upload.expired() {
				if !logged {
					log.Debugf("sweeping to clean out expired upload / download streams...")
					logged = true
				}
				log.Debugf("clearing out upload stream [%s]... it expired on %s", id, upload.expires)
				cancel[upload.id] = upload.writer
				delete(s.uploads, id)
			}
		}
		for id, download := range s.downloads {
			total += 1
			if download.expired() {
				if !logged {
					log.Debugf("sweeping to clean out expired upload / download streams...")
					logged = true
				}
				log.Debugf("clearing out download stream [%s]...", id)
				delete(s.downloads, id)
			}
		}
		s.lock.Unlock()

		if len(cancel) > 0 {
			log.Debugf("swept up: clearing out %d of %d streams", len(cancel), total)
			for id, wr := range cancel {
				log.Debugf("canceling upload stream [%s]...", id)
				wr.Close()
				wr.Cancel()
			}
			log.Debugf("canceled all expired streams.")
		}
	}
}
