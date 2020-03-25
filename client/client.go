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
	"os"
)

type ControlClient struct {
	URL string

	Username string
	Password string
}

type Client struct {
	URL string
}

func NewControlClient(url, username, password string) ControlClient {
	return ControlClient{
		URL:      url,
		Username: username,
		Password: password,
	}
}

func NewClient(url string) Client {
	return Client{
		URL: url,
	}
}

func (cc *ControlClient) StartUpload(path string) (*StreamInfo, error) {
	client := &http.Client{}
	var out StreamInfo

	requestBody, err := json.Marshal(map[string]string{
		"path": path,
	})
	if err != nil {
		return nil, err
	}

	reqURL := cc.URL + "/upload"
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cc.Username, cc.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}

	c := &Client{
		URL: cc.URL + "/upload/" + out.ID,
	}
	fmt.Println("URL: ", c.URL)
	return &out, nil
}

func (c *Client) Upload(id, token string, in *os.File) error {
	client := &http.Client{}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		var data UploadData
		data.Data = base64.StdEncoding.EncodeToString([]byte(scanner.Text()))
		requestBody, err := json.Marshal(data)
		if err != nil {
			return err
		}
		reqURL := c.URL + "/upload/" + id
		req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(requestBody))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-SSG-Token", token)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var out Response
		err = json.Unmarshal(b, &out)
		fmt.Println("OK: ", out.OK)
	}
	return nil
}

func (cc *ControlClient) StartDownload(path string) (*StreamInfo, error) {
	client := &http.Client{}
	var out StreamInfo

	requestBody, err := json.Marshal(map[string]string{
		"path": path,
	})
	if err != nil {
		return nil, err
	}

	reqURL := cc.URL + "/download"
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cc.Username, cc.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *Client) Download(id, token string) (io.ReadCloser, error) {
	client := &http.Client{}

	reqURL := c.URL + "/download/" + id
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SSG-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, err
}
