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
					log.Debugf(LOG + "sweeping to clean out expired upload / download streams...")
					logged = true
				}
				log.Debugf(LOG+"clearing out upload stream %v... it expired on %s", id, upload.expires)
				cancel[upload.id] = upload.writer
				upload.bucket.metrics.CancelUpload()
				delete(s.uploads, id)
			}
		}
		for id, download := range s.downloads {
			total += 1
			if download.expired() {
				if !logged {
					log.Debugf(LOG + "sweeping to clean out expired upload / download streams...")
					logged = true
				}
				log.Debugf(LOG+"clearing out download stream %v... it expired on %s", id, download.expires)
				download.bucket.metrics.CancelDownload()
				delete(s.downloads, id)
			}
		}
		s.lock.Unlock()

		if len(cancel) > 0 {
			log.Debugf(LOG+"swept up: clearing out %d of %d streams", len(cancel), total)
			for id, wr := range cancel {
				log.Debugf(LOG+"canceling upload stream %v...", id)
				wr.Close()
				wr.Cancel()
			}
			log.Debugf(LOG+"canceled all %d expired upload streams", len(cancel))
		}
	}
}
