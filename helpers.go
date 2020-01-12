package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	pb "github.com/adonese/microservices/raterpc/rate"
	"golang.org/x/net/html"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	address = "localhost:50051"
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
	// url := ip + "/" + endpoint
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
	log.Printf("the response code is: %d\n", res.StatusCode)
	if res.StatusCode == http.StatusOK {
		json.Unmarshal(body, &noebs)
		log.Printf("The passed Response object is: %+v\n", noebs.Response)
		return noebs.Response, nil
	}

	var ebsErr Error
	json.Unmarshal(body, &ebsErr)

	log.Printf("ebs response raw is: %s\n", string(body))

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

func parseBillers(billInfo map[string]interface{}) string {
	var m necBill
	//FIXME there is a bug here
	//mapFields, _ := additionalFieldsToHash(c.BillInfo)
	m.NewFromMap(billInfo)
	return m.Token
}

func (n *necBill) NewFromMap(f map[string]interface{}) {
	/*
	   "accountNo": "AM042111907231",
	   "customerName": "ALSAFIE BAKHIEYT HEMYDAN",
	   "meterFees": "0",
	   "meterNumber": "04203594959",
	   "netAmount": "10",
	   "opertorMessage": "Credit Purchase",
	   "token": "07246305192693082213",
	   "unitsInKWh": "66.7",
	   "waterFees": "0.00"
	*/
	n.SalesAmount, _ = strconv.ParseFloat(f["netAmount"].(string), 32)
	n.CustomerName = f["customerName"].(string)
	n.FixedFee, _ = strconv.ParseFloat(f["meterFees"].(string), 32)
	n.MeterNumber = f["meterNumber"].(string)
	n.Token = f["token"].(string)
}

const (
	zainTopUp	= "0010010001"
	zainBillPayment	= "0010010002"
	mtnTopUp	= "0010010003"
	mtnBillPayment= 	"0010010004"
	sudaniTopUp	= "0010010005"
	sudaniBillPayment = 	"0010010006"
	necPayment          = "0010020001"
	zainBillInquiry     = "0010010002"
	mtnBillInquiry      = "0010010004"
	sudaniBillInquiry   = "0010010006"
	moheBillInquiry     = "0010030002"
	moheBillPayment     = "0010030002"
	customsBillInquiry  = "0010030003"
	customsBillPayment  = "0010030003"
	moheArabBillInquiry = "0010030004"
	moheArabBillPayment = "0010030004"
	e15BillInquiry      = "0010050001"
	e15BillPayment      = "0010050001"
)

func additionalFieldsToHash(a string) (map[string]string, error) {
	fields := strings.Split(a, ";")
	if len(fields) < 2 {
		return nil, errors.New("index out of range")
	}
	out := make(map[string]string)
	for _, v := range fields {
		f := strings.Split(v, "=")
		out[f[0]] = f[1]
	}
	return out, nil
}

func necFormatter(bInfo map[string]interface{}) string {
	fullMessage := fmt.Sprintf("Token: %v\nUnits : %v [KW]\nCustomer Name: %v\nMeter Fees: %v\nWater Fees: %v\nNet Amount: %v\nMeter Number: %v\n\n**Thank You!**",
		bInfo["token"],
		bInfo["unitsInKWh"],
		bInfo["customerName"],
		bInfo["meterFees"],
		bInfo["waterFees"],
		bInfo["netAmount"],
		bInfo["meterNumber"])

	return fullMessage
}

func split(req string) []string {
	// req = strings.TrimRight(req, " ")
	f := strings.Split(req, " ")
	return f
}

func dispatch(f string) [][]string {
	f = normalize(f)
	s := split(f)
	length := len(s)
	var res [][]string

	upper := length - 2

	log.Printf("The upper val is: %v\n\n", upper)
	for i := 0; i <= upper; {
		v := split(f)
		res = append(res, []string{v[i], v[i+1]})
		i += 2
	}
	return res

}

func isOdd(d int) bool {
	return math.Mod(float64(d), 2) > 0

}

//toStrings normalized list val
func toStrings(l []string) string {
	return normalize(strings.Join(l, " "))
}

func getBiller(b string) (string, string) {
	//"MPHONE="
	pre := "MPHONE="
	if strings.HasPrefix(b, "092") || strings.HasPrefix(b, "099") {
		return mtnBillPayment, pre
	} else if strings.HasPrefix(b, "01") {
		return sudaniBillPayment, pre
	} else if strings.HasPrefix(b, "04") {
		return necPayment, "METER="
	}
	return zainBillInquiry, pre
}

func getTopUp(b string) (string, string) {
	//"MPHONE="
	pre := "MPHONE="
	if strings.HasPrefix(b, "092") || strings.HasPrefix(b, "099") {
		return mtnTopUp, pre
	} else if strings.HasPrefix(b, "01") {
		return sudaniTopUp, pre
	} else if strings.HasPrefix(b, "04") {
		return necPayment, "METER="
	}
	return zainTopUp, pre
}

func getInvoices(b string) string {
	if strings.HasPrefix(b, "092") || strings.HasPrefix(b, "099") {
		return mtnBillInquiry
	} else if strings.HasPrefix(b, "01") {
		return sudaniBillInquiry
	}
	return zainBillInquiry
}

func generateDate() string {
	y := time.Now().Year()
	M := time.Now().Month()
	d := time.Now().Day()
	h := time.Now().Hour()
	m := time.Now().Minute()
	s := time.Now().Second()

	yS := fmt.Sprintf("%d", y)
	return fmt.Sprintf("%s%02d%02d%02d%02d%02d", yS[2:], M, d, h, m, s)
}

func normalize(s string) string {
	s = strings.TrimSpace(s)
	return s
}

func logPanic(e error) {
	log.Printf("There is an error: %v", e)
}

func rpcClient() float32 {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRaterClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := c.GetSDGRate(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %f", r.Message)
	return r.Message

}
