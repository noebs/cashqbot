package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func request(buf []byte, url string) (Response, error) {

	verifyTLS := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	ebsClient := http.Client{
		Timeout:   30 * time.Second,
		Transport: verifyTLS,
	}

	log.Printf("The sent request is: %v\n\n", string(buf))
	reqBuilder, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(buf))

	reqBuilder.Header.Add("content-type", "application/json")
	res, err := ebsClient.Do(reqBuilder)
	if err != nil {
		log.Printf("The error is: %v", err)
		return Response{}, errors.New("it doesn't work")
	}

	body, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()

	log.Printf("The returned response (raw) is: %v", string(body))

	var noebs Noebs
	if res.StatusCode == http.StatusOK {
		json.Unmarshal(body, &noebs)
		log.Printf("The passed Response object is: %+v\n", noebs.Response)
		return noebs.Response, nil
	}
	var ebsErr Error
	err = json.Unmarshal(body, &ebsErr)
	if err != nil {
		log.Printf("The error is: %v", err)
		return ebsErr.Details, err
	}
	log.Printf("The passed response object is: %+v", ebsErr.Details)
	return ebsErr.Details, errors.New("there is something error")
}
