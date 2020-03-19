package api

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jhunt/go-route"
)

type API struct {
	FileRoot string
	Debug    bool
	Lease    time.Duration

	Control route.BasicAuth
	Admin   route.BasicAuth

	lock      sync.Mutex
	uploads   map[string]*Stream
	downloads map[string]*Stream
}

func New(path string) API {
	return API{
		FileRoot: path,

		uploads:   make(map[string]*Stream),
		downloads: make(map[string]*Stream),
	}
}

func (a *API) Debugf(f string, args ...interface{}) {
	if a.Debug {
		fmt.Fprintf(os.Stderr, f+"\n", args...)
	}
}

func (a *API) Sweeper(every time.Duration) {
	t := time.NewTicker(every)
	for range t.C {
		a.Sweep()
	}
}

func (a *API) Sweep() {
	a.lock.Lock()
	defer a.lock.Unlock()

	total := 0
	cleaned := 0
	a.Debugf("sweeping to clean out expired upload / download streams...")
	for id, s := range a.uploads {
		total += 1
		if s.Expired() {
			cleaned += 1
			// FIXME in S3 land this may block for a while; best
			// to move this Stream to a different queue, and let that
			// happen in a goroutine.
			a.Debugf("aborting upload stream [%s]...", id)
			s.Undo()
			a.Debugf("clearing out upload stream [%s]...", id)
			delete(a.uploads, id)
		}
	}

	for id, s := range a.downloads {
		total += 1
		if s.Expired() {
			cleaned += 1
			a.Debugf("clearing out download stream [%s]...", id)
			delete(a.downloads, id)
		}
	}

	a.Debugf("swept up.  cleared out %d of %d streams", cleaned, total)
}

func (a *API) NewUploadStream(path string) (*Stream, error) {
	a.Debugf("creating new upload stream for '%s'", path)
	s, err := NewStream(a.FileRoot+"/"+path, a.Lease)
	if err != nil {
		a.Debugf("failed to create new upload stream for '%s': %s", path, err)
		return nil, err
	}

	a.Debugf("persisting new upload stream as [%s]", s.ID)
	a.lock.Lock()
	defer a.lock.Unlock()
	a.uploads[s.ID] = &s
	return &s, nil
}

func (a *API) GetUploadStream(id string, token string) (*Stream, bool) {
	a.Debugf("retrieving upload stream [%s]...", id)
	a.lock.Lock()
	defer a.lock.Unlock()

	s, ok := a.uploads[id]
	if ok && s.Authorize(token) {
		s.Token.Renew()
		a.Debugf("renewing token for stream [%s] to %v", s.ID, s.Token.Expires)
		return s, true
	}
	return s, false
}

func (a *API) NewDownloadStream(path string) (*Stream, error) {
	a.Debugf("creating new download stream for '%s'", path)
	s, err := NewStream(a.FileRoot+"/"+path, a.Lease)
	if err != nil {
		a.Debugf("failed to create new download stream for '%s': %s", path, err)
		return nil, err
	}

	a.Debugf("persisting new download stream as [%s]", s.ID)
	a.lock.Lock()
	defer a.lock.Unlock()
	a.downloads[s.ID] = &s
	return &s, nil
}

func (a *API) GetDownloadStream(id string, token string) (*Stream, bool) {
	a.Debugf("retrieving download stream [%s]...", id)
	a.lock.Lock()
	defer a.lock.Unlock()

	s, ok := a.downloads[id]
	if ok && s.Authorize(token) {
		s.Token.Renew()
		a.Debugf("renewing token for stream [%s] to %v", s.ID, s.Token.Expires)
		return s, true
	}
	return s, false
}

func (a *API) ForgetUploadStream(s *Stream) {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.uploads, s.ID)
}

func (a *API) ForgetDownloadStream(s *Stream) {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.downloads, s.ID)
}

func (a *API) Router() *route.Router {
	r := &route.Router{}

	r.Dispatch("POST /download", func(r *route.Request) {
		if !r.BasicAuth(a.Control) {
			return
		}

		var in struct {
			Path  string `json:"path"`
			Agent string `json:"agent"`
		}
		if !r.Payload(&in) {
			return
		}
		if r.Missing("path", in.Path) || r.Missing("agent", in.Agent) {
			return
		}

		s, err := a.NewDownloadStream(in.Path)
		if err != nil {
			r.Fail(route.Oops(err, "Unable to create download stream"))
			return
		}

		r.OK(struct {
			ID      string    `json:"id"`
			Token   string    `json:"token"`
			Expires time.Time `json:"expires"`
		}{
			ID:      s.ID,
			Token:   s.Token.Secret,
			Expires: s.Token.Expires,
		})
	})

	r.Dispatch("POST /upload", func(r *route.Request) {
		if !r.BasicAuth(a.Control) {
			return
		}

		var in struct {
			Path  string `json:"path"`
			Agent string `json:"agent"`
		}
		if !r.Payload(&in) {
			return
		}
		if r.Missing("path", in.Path) || r.Missing("agent", in.Agent) {
			return
		}

		s, err := a.NewUploadStream(in.Path)
		if err != nil {
			r.Fail(route.Oops(err, "Unable to create download stream"))
			return
		}

		r.OK(struct {
			ID      string    `json:"id"`
			Token   string    `json:"token"`
			Expires time.Time `json:"expires"`
		}{
			ID:      s.ID,
			Token:   s.Token.Secret,
			Expires: s.Token.Expires,
		})
	})

	r.Dispatch("GET /download/:uuid", func(r *route.Request) {
		token := r.Req.Header.Get("X-SSG-Token")
		s, ok := a.GetDownloadStream(r.Args[1], token)
		if !ok {
			r.Fail(route.NotFound(nil, "stream not found"))
			return
		}

		out, err := s.Reader(token)
		if err != nil {
			r.Fail(route.Oops(err, "failed to read from download stream"))
			return
		}
		r.Header().Set("Content-Type", "application/octet-stream")
		r.Stream(out)
		a.ForgetDownloadStream(s)
	})

	r.Dispatch("POST /upload/:uuid", func(r *route.Request) {
		var in struct {
			Data string `json:"data"`
			EOF  bool   `json:"eof"`
		}

		if !r.Payload(&in) {
			return
		}

		token := r.Req.Header.Get("X-SSG-Token")
		s, ok := a.GetUploadStream(r.Args[1], token)
		if !ok {
			r.Fail(route.NotFound(nil, "stream not found"))
			return
		}

		if in.EOF {
			a.ForgetUploadStream(s)
			r.Success("upload finished")
			return
		}

		b, err := base64.StdEncoding.DecodeString(in.Data)
		if err != nil {
			r.Fail(route.Bad(err, "unable to decode base64 payload"))
			return
		}

		n, err := s.UploadChunk(token, b)
		if err != nil {
			r.Fail(route.Oops(err, "unable to upload data to stream"))
			return
		}
		r.Success("uploaded %d bytes", n)
	})

	r.Dispatch("GET /admin/streams", func(r *route.Request) {
		if !r.BasicAuth(a.Admin) {
			return
		}

		a.lock.Lock()
		defer a.lock.Unlock()

		type StreamInfo struct {
			ID       string    `json:"id"`
			Path     string    `json:"path"`
			Received uint64    `json:"recv"`
			Expires  time.Time `json:"expires"`
		}

		uploads := make([]StreamInfo, 0)
		for _, s := range a.uploads {
			uploads = append(uploads, StreamInfo{
				ID:       s.ID,
				Path:     s.Path,
				Received: s.Received,
				Expires:  s.Token.Expires,
			})
		}

		downloads := make([]StreamInfo, 0)
		for _, s := range a.downloads {
			downloads = append(downloads, StreamInfo{
				ID:      s.ID,
				Path:    s.Path,
				Expires: s.Token.Expires,
			})
		}

		r.OK(struct {
			Uploads   []StreamInfo `json:"uploads"`
			Downloads []StreamInfo `json:"downloads"`
		}{
			Uploads:   uploads,
			Downloads: downloads,
		})
	})

	return r
}
