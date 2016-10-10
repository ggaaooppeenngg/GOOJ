package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
)

var (
	bucket string
	domain string
)

func init() {

	conf.ACCESS_KEY = os.Getenv("QINIU_ACCESS_KEY")
	conf.SECRET_KEY = os.Getenv("QINIU_SECRET_KEY")
	bucket = os.Getenv("QINIU_BUCKET")
	domain = os.Getenv("QINIU_DOMAIN")

}

// ret type
type PutRet struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

func SaveFile(key string, content string) error {
	// This a a workaround, for test should not write real file
	// TODO: abstract fs interface

	if inTest {
		return nil
	}
	c := kodo.New(0, nil)
	policy := &kodo.PutPolicy{
		Scope: bucket,
	}
	var ret PutRet
	token := c.MakeUptoken(policy)
	uploader := kodocli.NewUploader(0, nil)
	err := uploader.Put(nil, &ret, token, key, strings.NewReader(content), int64(len(content)), nil)
	if err != nil {
		return err
	}
	return nil
}

// GetFIle gets a file, it is the caller's reponsibility to close file.
func GetFile(key string) (io.ReadCloser, error) {
	baseUrl := kodo.MakeBaseUrl(domain, key) // download url
	resp, err := http.Get(baseUrl)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("Status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
