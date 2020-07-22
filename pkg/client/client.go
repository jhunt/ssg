package client

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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

type Bucket struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Compression string `json:"compression"`
	Encryption  string `json:"encryption"`
}

type Client struct {
	URL          string
	ControlToken string
	SegmentSize  int

	Client *http.Client
}

func (c *Client) init() {
	if c.SegmentSize == 0 {
		c.SegmentSize = 1024 * 1024 // 1MiB
	}
	if c.Client == nil {
		c.Client = &http.Client{}
	}
}

func (c *Client) url(rest ...string) string {
	base := strings.TrimSuffix(c.URL, "/")
	return base + "/" + strings.Join(rest, "/")
}

func (c *Client) blob(id string) string {
	return c.url("blob", id)
}

func (c *Client) control(kind, target string) (*Stream, error) {
	c.init()

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

	req, err := http.NewRequest("POST", c.url("control"), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ControlToken)

	res, err := c.Client.Do(req)
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

func (c *Client) agent(id, token string, data []byte, eof bool) (int, error) {
	c.init()

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

	req, err := http.NewRequest("POST", c.blob(id), bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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

func (c *Client) Ping() (string, error) {
	c.init()

	req, err := http.NewRequest("GET", c.url(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	res, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("bad HTTP response %s", res.Status)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var out struct {
		Helo string `json:"ok"`
	}
	err = json.Unmarshal(b, &out)
	if err != nil {
		return "", err
	}
	return out.Helo, nil
}

func (c *Client) Buckets() ([]Bucket, error) {
	c.init()

	req, err := http.NewRequest("GET", c.url("buckets"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ControlToken)
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("bad HTTP response %s", res.Status)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var buckets []Bucket
	return buckets, json.Unmarshal(b, &buckets)
}

func (c *Client) NewUpload(target string) (*Stream, error) {
	return c.control("upload", target)
}

func (c *Client) NewDownload(target string) (*Stream, error) {
	return c.control("download", target)
}

func (c *Client) Expunge(target string) error {
	_, err := c.control("expunge", target)
	return err
}

func (c *Client) Put(id, token string, in io.Reader, eof bool) (int64, error) {
	c.init()

	var size int64

	scan := bufio.NewScanner(in)
	scan.Buffer(make([]byte, c.SegmentSize), c.SegmentSize)
	scan.Split(splitInto(c.SegmentSize))

	for scan.Scan() {
		n, err := c.agent(id, token, scan.Bytes(), false)
		if err != nil {
			return 0, err
		}
		size += int64(n)
	}

	if eof {
		_, err := c.agent(id, token, nil, true)
		if err != nil {
			return 1, err
		}
	}

	return size, nil
}

func (c *Client) Get(id, token string) (io.ReadCloser, error) {
	if c.Client == nil {
		c.Client = &http.Client{}
	}

	req, err := http.NewRequest("GET", c.blob(id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errorFrom(res)
	}

	return res.Body, err
}
