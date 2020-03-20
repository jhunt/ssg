package api

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jhunt/go-route"

	"github.com/shieldproject/shield-storage-gateway/backend"
)

type API struct {
	FileRoot string
	Debug    bool
	Lease    time.Duration

	Control route.BasicAuth
	Admin   route.BasicAuth

	builder  backend.BackendBuilder

	lock      sync.Mutex
	uploads   map[string]*Stream
	downloads map[string]*Stream
}

func New() API {
	return API{
		uploads:   make(map[string]*Stream),
		downloads: make(map[string]*Stream),
	}
}

func (a *API) UseFiles(root string) {
	a.builder = backend.FileBuilder(root)
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

			a.Debugf("canceling upload stream [%s]...", id)
			go s.Cancel()
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
	s, err := NewStream(path, a.builder)
	if err != nil {
		a.Debugf("failed to create new upload stream for '%s': %s", path, err)
		return nil, err
	}

	if err := s.Lease(a.Lease); err != nil {
		a.Debugf("failed to lease upload stream [%s]: %s", s.ID, err)
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
	return s, ok && s.Authorize(token)
}

func (a *API) ForgetUploadStream(s *Stream) {
	a.lock.Lock()
	defer a.lock.Unlock()
	delete(a.uploads, s.ID)
}

func (a *API) NewDownloadStream(path string) (*Stream, error) {
	a.Debugf("creating new download stream for '%s'", path)
	s, err := NewStream(path, a.builder)
	if err != nil {
		a.Debugf("failed to create new download stream for '%s': %s", path, err)
		return nil, err
	}

	if err := s.Lease(a.Lease); err != nil {
		a.Debugf("failed to lease download stream [%s]: %s", s.ID, err)
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
	return s, ok && s.Authorize(token)
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
		}
		if !r.Payload(&in) {
			return
		}
		if r.Missing("path", in.Path) {
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
			Token:   s.Token(),
			Expires: s.Expires(),
		})
	})

	r.Dispatch("POST /upload", func(r *route.Request) {
		if !r.BasicAuth(a.Control) {
			return
		}

		var in struct {
			Path  string `json:"path"`
		}
		if !r.Payload(&in) {
			return
		}
		if r.Missing("path", in.Path) {
			return
		}

		s, err := a.NewUploadStream(in.Path)
		if err != nil {
			r.Fail(route.Oops(err, "Unable to create upload stream"))
			return
		}

		r.OK(struct {
			ID      string    `json:"id"`
			Token   string    `json:"token"`
			Expires time.Time `json:"expires"`
		}{
			ID:      s.ID,
			Token:   s.Token(),
			Expires: s.Expires(),
		})
	})

	r.Dispatch("GET /download/:uuid", func(r *route.Request) {
		token := r.Req.Header.Get("X-SSG-Token")
		s, ok := a.GetDownloadStream(r.Args[1], token)
		if !ok {
			r.Fail(route.NotFound(nil, "stream not found"))
			return
		}

		out, err := s.AuthorizedRetrieve(token)
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

		n, err := s.AuthorizedWrite(token, b)
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
				Expires:  s.Expires(),
			})
		}

		downloads := make([]StreamInfo, 0)
		for _, s := range a.downloads {
			downloads = append(downloads, StreamInfo{
				ID:      s.ID,
				Path:    s.Path,
				Expires: s.Expires(),
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
