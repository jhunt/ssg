package client

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Customer struct {
	URL string

	Client *http.Client
}

func (c *Customer) url(id string) string {
	base := strings.TrimSuffix(c.URL, "/")
	return base + "/blob/" + id
}

type segment struct {
}

func (c *Customer) send(id, token string, data []byte, eof bool) (int, error) {
	if c.Client == nil {
		c.Client = &http.Client{}
	}

	var seg struct {
		Data string `json:"data"`
		EOF  bool   `json:"eof"`
	}

	if data != nil {
		seg.Data = base64.StdEncoding.EncodeToString(data)
	}
	seg.EOF = eof

	b, err := json.Marshal(seg)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", c.url(id), bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-SSG-Token", token)

	res, err := c.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, errorFrom(res)
	}

	return len(data), nil
}

func (c *Customer) Upload(id, token string, in io.Reader, eof bool) (int64, error) {
	var size int64

	scan := bufio.NewScanner(in)
	scan.Split(splitInto(8192))

	for scan.Scan() {
		n, err := c.send(id, token, scan.Bytes(), false)
		if err != nil {
			return 0, err
		}
		size += int64(n)
	}

	if eof {
		_, err := c.send(id, token, nil, true)
		if err != nil {
			return 1, err
		}
	}

	return size, nil
}

func (c *Customer) Download(id, token string) (io.ReadCloser, error) {
	if c.Client == nil {
		c.Client = &http.Client{}
	}

	req, err := http.NewRequest("GET", c.url(id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-SSG-Token", token)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errorFrom(res)
	}

	return res.Body, err
}
