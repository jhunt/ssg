package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type Controller struct {
	URL   string
	Token string

	Client *http.Client
}

func (cc *Controller) url(rest ...string) string {
	base := strings.TrimSuffix(cc.URL, "/")
	return base + "/" + strings.Join(rest, "/")
}

func (cc *Controller) control(kind, target string) (*Stream, error) {
	if cc.Client == nil {
		cc.Client = &http.Client{}
	}

	b, err := json.Marshal(struct {
		Kind   string `json:"kind"`
		Target string `json:"target"`
	}{
		Kind:   kind,
		Target: target,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", cc.url("control"), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cc.Token)

	res, err := cc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errorFrom(res)
	}

	defer res.Body.Close()
	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var out Stream
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (cc *Controller) NewUpload(target string) (*Stream, error) {
	return cc.control("upload", target)
}

func (cc *Controller) NewDownload(target string) (*Stream, error) {
	return cc.control("download", target)
}

func (cc *Controller) Expunge(target string) error {
	_, err := cc.control("expunge", target)
	return err
}
