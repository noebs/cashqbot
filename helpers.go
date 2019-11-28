package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
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
	defer res.Body.Close()

	var noebs Noebs
	if res.StatusCode == http.StatusOK {
		err := json.Unmarshal(body, &noebs)
		if err != nil {
			log.Printf("There is an error in noebser marshaling: %v", err)
		}
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

//extract extracts links of provided URL
func extract(domain string) []string {
	var links []string

	res, err := http.Get(domain)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		return nil
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "strong" {
			// for _, a := range n.Attr {
			// 	if a.Key == "span" {
			// 		fmt.Printf("The value we found is: %v", a)
			// 	}
			// }

			// if n.FirstChild.Data == "سعر الدولار الأمريكي" {
			// 	fmt.Printf("The value we want is: %v\n%v", n.FirstChild.NextSibling.Data, n.FirstChild.Data)
			// }
			links = append(links, n.FirstChild.Data)
			// fmt.Printf("The values are: %#v\n", n.FirstChild.Data)
			// for _, s := range n.Attr {
			// 	fmt.Printf("The value is :%v\n", s.Val)
			// }

		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	//
	return links
}

func dump(links []string) (bool, string) {
	for i, v := range links {
		if v == "سعر الدولار الأمريكي" {
			usd := strings.Split(links[i+1], " ")
			return true, usd[0]
		}
	}
	return false, ""
}
