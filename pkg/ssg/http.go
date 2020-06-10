package ssg

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/jhunt/go-log"
	"github.com/jhunt/go-route"

	"github.com/jhunt/shield-storage-gateway/pkg/url"
)

func (s *Server) Router(helo string) *route.Router {
	r := &route.Router{}

	authz := func(r *route.Request, allowed []string) bool {
		tok := r.Req.Header.Get("Authorization")
		if tok == "" {
			r.Fail(route.Unauthorized(nil, "control auth required"))
			return false
		}

		if !strings.HasPrefix(tok, "Bearer ") {
			r.Fail(route.Forbidden(nil, "forbidden"))
			return false
		}

		tok = strings.TrimPrefix(tok, "Bearer ")

		for i := range allowed {
			if allowed[i] == tok {
				return true
			}
		}

		r.Fail(route.Forbidden(nil, "forbidden"))
		return false
	}

	r.Dispatch("GET /", func(r *route.Request) {
		r.Success(helo)
	})

	r.Dispatch("POST /control", func(r *route.Request) {
		if !authz(r, s.ControlTokens) {
			return
		}

		var in struct {
			Kind   string `json:"kind"`
			Target string `json:"target"`
		}
		if !r.Payload(&in) {
			return
		}
		if r.Missing("kind", in.Kind) {
			return
		}

		if in.Kind != "upload" && in.Kind != "download" && in.Kind != "expunge" {
			r.Fail(route.Bad(nil, "invalid kind: '%s'", in.Kind))
			return
		}

		target, err := url.Parse(in.Target)
		if err != nil {
			r.Fail(route.Bad(err, "invalid target '%s': %s", in.Target, err))
			return
		}

		switch in.Kind {
		case "upload":
			stream, path, err := s.startUpload(target)
			if err != nil {
				r.Fail(route.Oops(err, "unable to start upload"))
				return
			}

			target.Path = path
			r.OK(struct {
				Kind    string    `json:"kind"`
				ID      string    `json:"id"`
				Token   string    `json:"token"`
				Canon   string    `json:"canon"`
				Expires time.Time `json:"expires"`
			}{
				Kind:    "upload",
				ID:      stream.id,
				Token:   stream.secret,
				Canon:   target.String(),
				Expires: stream.expires,
			})
			return

		case "download":
			stream, err := s.startDownload(target)
			if err != nil {
				r.Fail(route.Oops(err, "unable to start download"))
				return
			}

			target.Cluster = s.Cluster
			r.OK(struct {
				Kind    string    `json:"kind"`
				ID      string    `json:"id"`
				Token   string    `json:"token"`
				Canon   string    `json:"canon"`
				Expires time.Time `json:"expires"`
			}{
				Kind:    "download",
				ID:      stream.id,
				Token:   stream.secret,
				Canon:   target.String(),
				Expires: stream.expires,
			})
			return

		case "expunge":
			err := s.expunge(target)
			if err != nil {
				r.Fail(route.Oops(err, "unable to expunge"))
				return
			}

			r.OK(struct {
				Kind  string `json:"kind"`
				Canon string `json:"canon"`
			}{
				Kind:  "expunge",
				Canon: target.String(),
			})
			return
		}
	})

	r.Dispatch("GET /blob/:id", func(r *route.Request) {
		token := r.Req.Header.Get("X-SSG-Token")
		downstream, ok := s.getDownload(r.Args[1], token)
		if !ok {
			r.Fail(route.NotFound(nil, "stream not found"))
			return
		}

		r.Header().Set("Content-Type", "application/octet-stream")
		r.Stream(downstream)
		downstream.Close()
		s.forget(downstream)
	})

	r.Dispatch("POST /blob/:id", func(r *route.Request) {
		token := r.Req.Header.Get("X-SSG-Token")
		upstream, ok := s.getUpload(r.Args[1], token)
		if !ok {
			r.Fail(route.NotFound(nil, "stream not found"))
			return
		}

		var in struct {
			Data string `json:"data"`
			EOF  bool   `json:"eof"`
		}
		if !r.Payload(&in) {
			return
		}

		n := 0
		if in.Data != "" {
			b, err := base64.StdEncoding.DecodeString(in.Data)
			if err != nil {
				r.Fail(route.Bad(err, "unable to decode base64 payload"))
				return
			}

			log.Debugf(LOG+"uploading %d bytes (eof: %v) to stream %v", len(b), in.EOF, upstream.id)
			n, err = upstream.Write(b)
			if err != nil {
				r.Fail(route.Oops(err, "unable to upload data to stream"))
				return
			}
		}

		if in.EOF {
			log.Debugf(LOG+"EOF signaled by client; closing upload stream %v", upstream.id)
			upstream.Close()
			s.forget(upstream)
		}

		r.OK(struct {
			Segments     int   `json:"segments"`
			Compressed   int64 `json:"compressed"`
			Uncompressed int64 `json:"uncompressed"`
			Sent         int   `json:"sent",omitempty`
		}{
			Segments:     upstream.segments,
			Compressed:   upstream.compressed.total(),
			Uncompressed: upstream.uncompressed.total(),
			Sent:         n,
		})
	})

	r.Dispatch("GET /streams", func(r *route.Request) {
		if !authz(r, s.ControlTokens) {
			return
		}

		type info struct {
			Kind     string    `json:"kind"`
			ID       string    `json:"id"`
			Canon    string    `json:"canon"`
			Expires  time.Time `json:"expires"`
			Received int64     `json:"received"`
		}

		s.lock.Lock()
		defer s.lock.Unlock()
		l := make([]info, 0)
		for _, v := range s.uploads {
			l = append(l, info{
				Kind:     "upload",
				ID:       v.id,
				Canon:    v.canon,
				Expires:  v.expires,
				Received: v.uncompressed.total(),
			})
		}
		for id, v := range s.downloads {
			l = append(l, info{
				Kind:     "download",
				ID:       id,
				Canon:    v.canon,
				Expires:  v.expires,
				Received: v.uncompressed.total(),
			})
		}
		r.OK(l)
	})

	r.Dispatch("DELETE /streams/:id", func(r *route.Request) {
		if !authz(r, s.ControlTokens) {
			return
		}
	})

	r.Dispatch("GET /metrics", func(r *route.Request) {
		if !authz(r, s.MonitorTokens) {
			return
		}

		m := make(map[string]metrics)
		for _, b := range s.buckets {
			b.metrics.Recalculate()
			m[b.key] = b.metrics
		}
		r.OK(m)
	})

	r.Dispatch("DELETE /metrics", func(r *route.Request) {
		if !authz(r, s.MonitorTokens) {
			return
		}

		m := make(map[string]metrics)
		for _, b := range s.buckets {
			b.metrics.Reset()
			m[b.key] = b.metrics
		}
		r.OK(m)
	})

	return r
}
