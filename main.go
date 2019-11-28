package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	zain   = "0010010001"
	mtn    = "0010010003"
	sudani = "0010010005"
	nec    = "0010020001"
)

func main() {
	c := cache.New(5*time.Hour, 10*time.Hour)

	a := extract("https://www.price-today.com/currency-prices-sudan/")
	fmt.Printf("The values are: %v\n", a)
	_, r := dump(a)
	c.Set("rate", r, 24*time.Hour)

	b, err := tb.NewBot(tb.Settings{
		Token: "1001304778:AAGqNz-9ESmnMjMcsIqzsN_1A_ncWydb6fw",
		// You can also set custom API URL. If field is empty it equals to "https://api.telegram.org"
		//URL:    "http://195.129.111.17:8012",
		Poller: &tb.LongPoller{Timeout: 2 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/help", func(m *tb.Message) {
		h := `Help: CashQ bot is our friendly bot that can help you to do your payments!
		
		List of commands (not commands start with a "/" prefix)

		/rate (our most useful command)
		rate helps you find the accurate rate of US dollar to SDG. We use the
		service provided by for our
		pricing.

		/balance PAN IPIN Expdate
		- Enter your PAN (16 or 19 digit)
		- Your IPIN (note, internet PIN and not your typical PIN)
		- Expdate (as written in your Cards YYMM or Last two digits of year and month 01)
		
		/zain PAN IPIN Expdate MobileNumber Amount
		- Enter your PAN (16 or 19 digit)
		- Your IPIN (note, internet PIN and not your typical PIN)
		- Expdate (as written in your Cards YYMM or Last two digits of year and month 01)
		- Mobile number: Your mobile number
		- Amount: The amount for top up
		
		/sudani PAN IPIN Expdate MobileNumber Amount
		- Enter your PAN (16 or 19 digit)
		- Your IPIN (note, internet PIN and not your typical PIN)
		- Expdate (as written in your Cards YYMM or Last two digits of year and month 01)
		- Mobile number: Your mobile number
		- Amount: The amount for top up
		
		/mtn PAN IPIN Expdate MobileNumber Amount
		- Enter your PAN (16 or 19 digit)
		- Your IPIN (note, internet PIN and not your typical PIN)
		- Expdate (as written in your Cards YYMM or Last two digits of year and month 01)
		- Mobile number: Your mobile number
		- Amount: The amount for top up
		
		/nec PAN IPIN Expdate NEC_Number Amount
		- Enter your PAN (16 or 19 digit)
		- Your IPIN (note, internet PIN and not your typical PIN)
		- Expdate (as written in your Cards YYMM or Last two digits of year and month 01)
		- NEC_number: Your NEC number (11 digits)
		- Amount: The amount for electricity payment
		
		/p2p PAN IPIN Expdate ToCard Amount
		- Enter your PAN (16 or 19 digit)
		- Your IPIN (note, internet PIN and not your typical PIN)
		- Expdate (as written in your Cards YYMM or Last two digits of year and month 01)
		- ToCard: The 16-19 digits of card you want to send to
		- Amount: The amount for card transfer
		
		Built with <3 by your friends at Solus https://soluspay.net
		*Download Our Android App CashQ for full mobile experience https://play.google.com/store/apps/details?id=net.soluspay.cashq`

		replyBtn := tb.ReplyButton{
			Text: "Download Cashq 📢",
		}
		visitSolus := tb.ReplyButton{
			Text: "Visit Solus <3",
		}

		replyKeys := [][]tb.ReplyButton{
			[]tb.ReplyButton{replyBtn},
			[]tb.ReplyButton{visitSolus},
		}

		b.Handle(&replyBtn, func(m *tb.Message) {
			// how to do something?

		})
		b.Send(m.Sender, h, &tb.ReplyMarkup{
			ReplyKeyboard: replyKeys,
		})
	})

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, m.Payload)
	})

	b.Handle("/rate", func(m *tb.Message) {
		if res, ok := c.Get("rate"); !ok {
			a := extract("https://www.price-today.com/currency-prices-sudan/")
			fmt.Printf("The values are: %v\n", a)
			_, r := dump(a)
			fmt.Printf("The USD rate is: %v\n", r)
			c.Set("rate", r, 24*time.Hour)
		} else {
			b.Send(m.Sender, fmt.Sprintf("The rate for USD is: %vSDG\nThanks Hamadok 😘📢", res.(string)))
		}

	})

	b.Handle("/balance", func(m *tb.Message) {
		// get key
		payload := m.Payload
		p := strings.Split(payload, " ")
		fmt.Printf("The payload is: %v", p)
		if len(p) < 3 {
			b.Send(m.Sender, "Please send: PAN, IPIN, and ExpDate")
			return
		}
		key, _ := getKey()
		uuid := uuid.New().String()
		ipin, _ := rsaEncrypt(key, p[1], uuid)
		res, err := balance(ipin, p[0], p[2], uuid)
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("EBS error: Response Message: %v\n🙏🙏🙏", res.ResponseMessage))
			return
		}

		b.Send(m.Sender, fmt.Sprintf("Your balance is: %v 💰\nMade with ❤ by your friends at Solus!", res.Balance["available"]))
		log.Printf("The balance is: %v", res.Balance["available"])
	})

	b.Handle("/zain", func(m *tb.Message) {
		// get key
		payload := m.Payload
		p := strings.Split(payload, " ")
		fmt.Printf("The payload is: %v", p)
		if len(p) < 5 {
			b.Send(m.Sender, "Please send: PAN, IPIN, and ExpDate, mobile number and amount")
			return
		}

		key, err := getKey()
		if err != nil {
			log.Printf("The erorr is: %v. The key is: %v\n", err, key)
			b.Send(m.Sender, "Failed to process the transaction. Code PUB_KEY_ERR\n")
			return
		}
		uuid := uuid.New().String()
		ipin, err := rsaEncrypt(key, p[1], uuid)

		if err != nil {
			log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
			b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
			return
		}

		payInfo := "MPHONE=" + p[3]
		expDate := p[2]
		amount := p[4]
		pan := p[0]

		amountVal, _ := strconv.ParseFloat(amount, 32)
		res, err := billers(zain, payInfo, pan, ipin, expDate, uuid, amountVal)
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Codeode: %v", res.ResponseMessage, res.ResponseCode))
		} else {
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction. Reference number: %v", res.ResponseMessage))
		}

	})

	b.Handle("/sudani", func(m *tb.Message) {
		// get key
		payload := m.Payload
		p := strings.Split(payload, " ")
		fmt.Printf("The payload is: %v", p)
		if len(p) < 5 {
			b.Send(m.Sender, "Please send: PAN, IPIN, and ExpDate, mobile number and amount")
			return
		}

		key, err := getKey()
		if err != nil {
			log.Printf("The erorr is: %v. The key is: %v\n", err, key)
			b.Send(m.Sender, "Failed to process the transaction. Code PUB_KEY_ERR\n")
			return
		}
		uuid := uuid.New().String()
		ipin, err := rsaEncrypt(key, p[1], uuid)

		if err != nil {
			log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
			b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
			return
		}

		payInfo := "MPHONE=" + p[3]
		expDate := p[2]
		amount := p[4]
		pan := p[0]

		amountVal, _ := strconv.ParseFloat(amount, 32)
		res, err := billers(sudani, payInfo, pan, ipin, expDate, uuid, amountVal)
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("There is an error: %v. EBS response: %v", err, res.ResponseMessage))
		} else {
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction. Reference number: %v", res.ResponseMessage))
		}

	})

	b.Handle("/mtn", func(m *tb.Message) {
		// get key
		payload := m.Payload
		p := strings.Split(payload, " ")
		fmt.Printf("The payload is: %v", p)
		if len(p) < 5 {
			b.Send(m.Sender, "Please send: PAN, IPIN, and ExpDate, mobile number and amount")
			return
		}

		key, err := getKey()
		if err != nil {
			log.Printf("The erorr is: %v. The key is: %v\n", err, key)
			b.Send(m.Sender, "Failed to process the transaction. Code PUB_KEY_ERR\n")
			return
		}
		uuid := uuid.New().String()
		ipin, err := rsaEncrypt(key, p[1], uuid)

		if err != nil {
			log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
			b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
			return
		}

		payInfo := "MPHONE=" + p[3]
		expDate := p[2]
		amount := p[4]
		pan := p[0]

		amountVal, _ := strconv.ParseFloat(amount, 32)
		res, err := billers(mtn, payInfo, pan, ipin, expDate, uuid, amountVal)
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Codeode: %v", res.ResponseMessage, res.ResponseCode))
		} else {
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction. Reference number: %v", res.ResponseMessage))
		}

	})

	b.Handle("/nec", func(m *tb.Message) {
		// get key
		payload := m.Payload
		p := strings.Split(payload, " ")
		fmt.Printf("The payload is: %v", p)
		if len(p) < 5 {
			b.Send(m.Sender, "Please send: PAN, IPIN, and ExpDate, mobile number and amount")
			return
		}

		key, err := getKey()
		if err != nil {
			log.Printf("The erorr is: %v. The key is: %v\n", err, key)
			b.Send(m.Sender, "Failed to process the transaction. Code PUB_KEY_ERR\n")
			return
		}
		uuid := uuid.New().String()
		ipin, err := rsaEncrypt(key, p[1], uuid)

		if err != nil {
			log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
			b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
			return
		}

		payInfo := "METER=" + p[3]
		expDate := p[2]
		amount := p[4]
		pan := p[0]

		amountVal, _ := strconv.ParseFloat(amount, 32)
		res, err := billers(nec, payInfo, pan, ipin, expDate, uuid, amountVal)
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Codeode: %v", res.ResponseMessage, res.ResponseCode))
		} else {
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction. \nResponse Message: %v\nResponse Code is: %v, Token: %v", res.ResponseMessage, res.ResponseCode, res.BillInfo))
		}

	})

	b.Handle("/p2p", func(m *tb.Message) {
		// get key
		payload := m.Payload
		p := strings.Split(payload, " ")
		fmt.Printf("The payload is: %v", p)
		if len(p) < 5 {
			b.Send(m.Sender, "Please send: PAN, IPIN, and ExpDate, mobile number and amount")
			return
		}

		key, err := getKey()
		if err != nil {
			log.Printf("The erorr is: %v. The key is: %v\n", err, key)
			b.Send(m.Sender, "Failed to process the transaction. Code PUB_KEY_ERR\n")
			return
		}
		uuid := uuid.New().String()
		ipin, err := rsaEncrypt(key, p[1], uuid)

		if err != nil {
			log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
			b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
			return
		}

		toCard := p[3]
		expDate := p[2]
		amount := p[4]
		pan := p[0]

		amountVal, _ := strconv.ParseFloat(amount, 32)
		res, err := cardTransfer(toCard, pan, ipin, expDate, uuid, amountVal)
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Codeode: %v", res.ResponseMessage, res.ResponseCode))
		} else {
			log.Printf("The response body is: %v", res)
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction.\nResponse Message: %v\nResponse code: %v", res.ResponseMessage, res.ResponseCode))
		}

	})

	// experimental -- for performance
	go func() {
		<-time.After(60 * time.Second)
		b.Stop()
	}()

	b.Handle(tb.OnText, "unknown command")
	b.Start()
}

func getKey() (string, error) {
	k := Key{
		ApplicationId: "ACTSCon",
		TranDateTime:  "191124141930",
		UUID:          uuid.New().String(),
	}

	buf, _ := json.Marshal(&k)

	noebs, err := request(buf, "http://192.168.20.20:8080/consumer/key")
	if err != nil {
		return "", err
	}
	return noebs.Key, nil
}

func balance(ipin, pan, expDate, uuid string) (Response, error) {

	k := Balance{
		Key: Key{
			ApplicationId: "ACTSCon",
			TranDateTime:  "191124141930",
			UUID:          uuid,
		},
		Card: Card{
			PAN:     pan,
			IPIN:    ipin,
			Expdate: expDate,
		},
	}

	buf, _ := json.Marshal(&k)

	res, err := request(buf, "http://192.168.20.20:8080/consumer/balance")
	if err != nil {
		return res, err
	}
	return res, nil
}

// billers(nec, payInfo, pan, ipin, expDate, uuid, amount)
func billers(payeeId, personalInfo, pan, ipin, expDate, uuid string, amount float64) (Response, error) {

	/*
		zain top up: 0010010001
		mtn top up: 0010010003
		sudani top up: 0010010005
		nec top up: 0010020001
	*/

	var pId string

	switch p := payeeId; {
	case p == zain:
		pId = zain
	case p == mtn:
		pId = mtn
	case p == sudani:
		pId = sudani
	case p == nec:
		pId = nec
	}

	k := Bills{
		Key: Key{
			ApplicationId: "ACTSCon",
			TranDateTime:  "191124141930",
			UUID:          uuid,
		},
		Card: Card{
			PAN:     pan,
			Expdate: expDate,
			IPIN:    ipin,
		},
		PayeeId:      pId,
		PersonalInfo: personalInfo,
		Amount: Amount{
			AmountNumber:     amount,
			TranCurrencyCode: "SDG",
		},
	}

	buf, _ := json.Marshal(&k)

	noebs, err := request(buf, "http://192.168.20.20:8080/consumer/bill_payment")
	if err != nil {
		return noebs, err
	}
	return noebs, nil
}

func cardTransfer(toCard, pan, ipin, expDate, uuid string, amount float64) (Response, error) {

	/*
		zain top up: 0010010001
		mtn top up: 0010010003
		sudani top up: 0010010005
		nec top up: 0010020001
	*/

	k := CardTransfer{
		Key: Key{
			ApplicationId: "ACTSCon",
			TranDateTime:  "191124141930",
			UUID:          uuid,
		},
		Card: Card{
			PAN:     pan,
			Expdate: expDate,
			IPIN:    ipin,
		},
		ToCard: toCard,
		Amount: Amount{
			AmountNumber:     amount,
			TranCurrencyCode: "SDG",
		},
	}

	buf, _ := json.Marshal(&k)

	noebs, err := request(buf, "http://192.168.20.20:8080/consumer/p2p")
	if err != nil {
		return noebs, err
	}
	return noebs, nil
}

func getCardPayload(payload string) (Card, error) {
	p := strings.Split(payload, " ")
	fmt.Printf("The payload is: %v", p)
	if len(p) != 5 {
		return Card{}, errors.New("please send the right request")
	}
	c := Card{PAN: p[0], IPIN: p[1], Expdate: p[2]}
	return c, nil
}
