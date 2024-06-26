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

// ServerIP is used for our internal server
// curl https://100.74.147.70/api/isAlive
const ServerIP = "https://beta.soluspay.net/api/consumer/"

// currentPrice := make(chan float32)
var ticker = time.NewTicker(10 * time.Second)
var usdRate = rpcClient()
var c = cache.New(5*time.Hour, 5*time.Minute)

const (
	zain   = "0010010001"
	mtn    = "0010010003"
	sudani = "0010010005"
	nec    = "0010020001"
)

func main() {

	go tickerHandler()

	// a := extract("https://www.price-today.com/currency-prices-sudan/")
	// fmt.Printf("The values are: %v\n", a)
	// _, r := dump(a)

	c.Set("rate", usdRate, 24*time.Hour)

	b, err := tb.NewBot(tb.Settings{
		Token: "1001304778:AAGqNz-9ESmnMjMcsIqzsN_1A_ncWydb6fw",
		// You can also set custom API URL. If field is empty it equals to "https://api.telegram.org"
		//URL:    "http://195.129.111.17:8012",
		Poller:   &tb.LongPoller{Timeout: 2 * time.Second},
		Reporter: logPanic,
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
		
		/bills PAN IPIN Expdate [mobile_number, amount]   
		NOTE THIS API Supports bulk transactions. WHICH IS WAY COOLER THAN THE COOL ITSELF.
		It also automatically detects your telecos provider, thanks me!

		Ex for bulk transactions:
		/bills 92202121212121 1124 2203 09123456782 100 091323232 120 091323232232 10
		(it also supports mixins, i.e., pay to Zain and Sudani at the same time!)

		/invoices PAN IPIN Expdate [mobile_number, amount]   
		NOTE THIS API Supports bulk transactions. WHICH IS WAY COOLER THAN THE COOL ITSELF.
		It also automatically detects your telecos provider, thanks me!
		Returns the unbilled amount (the number you are supposed to pay)

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

		// replyBtn := tb.ReplyButton{
		// 	Text: "Download Cashq 📢",
		// }
		// visitSolus := tb.ReplyButton{
		// 	Text: "Visit Solus <3",
		// }

		// replyKeys := [][]tb.ReplyButton{
		// 	[]tb.ReplyButton{replyBtn},
		// 	[]tb.ReplyButton{visitSolus},
		// }

		// b.Handle(&replyBtn, func(m *tb.Message) {
		// 	// how to do something?

		// })
		b.Send(m.Sender, h)
	})

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, m.Payload)
	})

	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Sender, `Welcome to cashqbot your friendly payment 🤖
press /help for helping using cashqbot`)
	})

	b.Handle("/test", func(m *tb.Message) {

		var pin, pan string
		replyBtn := tb.ReplyButton{
			Text: "Enter PAN 📢",
		}
		visitSolus := tb.ReplyButton{
			Text: "Enter PIN",
		}

		replyKeys := [][]tb.ReplyButton{
			{replyBtn},
			{visitSolus},
		}

		b.Handle(&replyBtn, func(m *tb.Message) {
			// how to do something?
			pan = m.Payload

		})
		b.Handle(&visitSolus, func(m *tb.Message) {
			// how to do something?
			pin = m.Payload

		})
		b.Send(m.Sender, "Please enter stuff in the keyboard", &tb.ReplyMarkup{
			ReplyKeyboard: replyKeys,
		})
		// get pan and pin
		b.Send(m.Sender, fmt.Sprintf("Your pan is: %s, your pin is: %s\n", pan, pin))
	})

	b.Handle("/rate", func(m *tb.Message) {
		// if res, ok := c.Get("rate"); !ok {
		// 	a := extract("https://www.price-today.com/currency-prices-sudan/")
		// 	fmt.Printf("The values are: %v\n", a)
		// 	_, r := dump(a)
		// 	fmt.Printf("The USD rate is: %v\n", r)
		// 	c.Set("rate", r, 24*time.Hour)
		// } else {
		// 	b.Send(m.Sender, fmt.Sprintf("The rate for USD is: %vSDG\nThanks Hamadok 😘📢", res.(string)))
		// }
		// res, _ := c.Get("rate")
		b.Send(m.Sender, fmt.Sprintf("The rate for USD is: %vSDG\nThanks Hamadok 😘📢", usdRate))
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

		if time.Since(m.Time()) > time.Minute {
			b.Send(m.Sender, "Message time out. Try again 🙏")
			return
		}
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

		res, err := billers(true, zain, payInfo, pan, ipin, expDate, uuid, float32(amountVal))
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Code: %v", res.ResponseMessage, res.ResponseCode))
		} else {
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction. Reference number: %v", res.ResponseMessage))
		}

	})

	b.Handle("/sudani", func(m *tb.Message) {
		// get key
		if time.Since(m.Time()) > time.Minute {
			b.Send(m.Sender, "Message time out. Try again 🙏")
			return
		}

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
		res, err := billers(true, sudani, payInfo, pan, ipin, expDate, uuid, float32(amountVal))
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("There is an error: %v. EBS response: %v", err, res.ResponseMessage))
		} else {
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction. Reference number: %v", res.ResponseMessage))
		}

	})

	b.Handle("/mtn", func(m *tb.Message) {
		// get key
		if time.Since(m.Time()) > time.Minute {
			b.Send(m.Sender, "Message time out. Try again 🙏")
			return
		}

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
		res, err := billers(true, mtn, payInfo, pan, ipin, expDate, uuid, float32(amountVal))
		if err != nil {
			fmt.Printf("The error is: %v", err)
			b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Code: %v", res.ResponseMessage, res.ResponseCode))
		} else {
			b.Send(m.Sender, fmt.Sprintf("Successful Transaction. Reference number: %v", res.ResponseMessage))
		}

	})

	b.Handle("/nec", func(m *tb.Message) {
		// get key

		if time.Since(m.Time()) > time.Minute {
			b.Send(m.Sender, "Message time out. Try again 🙏")
			return
		}

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

		expDate := p[2]
		pan := p[0]

		// payInfo := "METER=" + p[3]
		// amount := p[4]

		data := toStrings(p[3:])
		log.Printf("The len of data is: %v", len(data))

		if isOdd(len(p[3:])) {
			// return an error
			// it should be a pari of nec_number, amount
			b.Send(m.Sender, "Wrong format. Please use nec_number amount")
			return
		}

		fields := dispatch(data)
		log.Printf("The fields output is: %#v", fields)
		var errCounter int

		for _, v := range fields {
			log.Printf("The bulked data is: %v, Type: %T", v[0], v)
			amountVal, _ := strconv.ParseFloat(v[1], 32)

			// generate ipin and generate uuid
			uuid := uuid.New().String()
			ipin, err := rsaEncrypt(key, p[1], uuid)

			if err != nil {
				log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
				b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
				return
			}

			res, err := billers(true, nec, "METER="+v[0], pan, ipin, expDate, uuid, float32(amountVal))
			if err != nil {
				fmt.Printf("The error is: %v", err)
				b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Code: %v", res.ResponseMessage, res.ResponseCode))
			} else {
				bInfo := res.BillInfo
				token := res.BillInfo["token"]
				fullMessage := necFormatter(bInfo)
				b.Send(m.Sender, fmt.Sprintf("Successful Transaction\nResponse Message: %v\nToken: ⚡%v⚡\n\n%v",
					res.ResponseMessage, token, fullMessage),
					&tb.SendOptions{
						ParseMode: "markdown",
					})
			}

		}

		// Count the Sum of the transactions.
		if errCounter > 0 {
			b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v\nAll Transaction(s) Completed with %d Errors\nMade with <3 By your friends at Solus!", time.Now().UTC(), errCounter))
			return
		}
		b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v, All Transaction(s) Completed Successfully\nMade with <3 By your friends at Solus!", time.Now().UTC()))

	})

	b.Handle("/p2p", func(m *tb.Message) {
		// get key

		if time.Since(m.Time()) > time.Minute {
			b.Send(m.Sender, "Message time out. Try again 🙏")
			return
		}

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

		if err != nil {
			log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
			b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
			return
		}

		expDate := p[2]
		pan := p[0]

		data := toStrings(p[3:])
		log.Printf("The len of data is: %v", len(data))

		if isOdd(len(p[3:])) {
			// return an error
			// it should be a pari of nec_number, amount
			b.Send(m.Sender, "Wrong format. Please use nec_number amount")
			return
		}

		fields := dispatch(data)
		log.Printf("The fields output is: %#v", fields)

		var errCounter int

		for _, v := range fields {
			log.Printf("The bulked data is: %v, Type: %T", v[0], v)
			amountVal, _ := strconv.ParseFloat(v[1], 32)
			// amountVal := fmt.Sprintf("%.2f", fval)

			// generate ipin and generate uuid
			uuid := uuid.New().String()
			ipin, err := rsaEncrypt(key, p[1], uuid)

			if err != nil {
				log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
				b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
				return
			}

			res, err := cardTransfer(v[0], pan, ipin, expDate, uuid, float32(amountVal))
			if err != nil {
				fmt.Printf("The error is: %v", err)
				b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Code: %v", res.ResponseMessage, res.ResponseCode))
			} else {
				b.Send(m.Sender, fmt.Sprintf("Successful Transaction\nResponse Status: %v\nResponse Message: %v",
					res.ResponseStatus, res.ResponseMessage))
			}

		}

		// Count the Sum of the transactions.
		if errCounter > 0 {
			b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v\nAll Transaction(s) Completed with %d Errors\nMade with <3 By your friends at Solus!", time.Now().UTC(), errCounter))
			return
		}
		b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v, All Transaction(s) Completed Successfully\nMade with <3 By your friends at Solus!", time.Now().UTC()))

	})

	b.Handle("/bills", func(m *tb.Message) {
		// get key

		if time.Since(m.Time()) > time.Minute {
			b.Send(m.Sender, "Message time out. Try again 🙏")
			return
		}

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

		expDate := p[2]
		pan := p[0]

		// payInfo := "METER=" + p[3]
		// amount := p[4]

		data := toStrings(p[3:])
		log.Printf("The len of data is: %v", len(data))

		if isOdd(len(p[3:])) {
			// return an error
			// it should be a pari of nec_number, amount
			b.Send(m.Sender, "Wrong format. Please use nec_number amount")
			return
		}

		fields := dispatch(data)
		log.Printf("The fields output is: %#v", fields)
		var errCounter int

		for _, v := range fields {
			log.Printf("The bulked data is: %v, Type: %T", v[0], v)
			amountVal, _ := strconv.ParseFloat(v[1], 32)
			// amountVal := fmt.Sprintf("%.2f", fval)

			// generate ipin and generate uuid
			uuid := uuid.New().String()
			ipin, err := rsaEncrypt(key, p[1], uuid)

			if err != nil {
				log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
				b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
				return
			}
			biller, pre := getBiller(v[0])
			res, err := billers(true, biller, pre+v[0], pan, ipin, expDate, uuid, float32(amountVal))
			if err != nil {
				errCounter++
				fmt.Printf("The error is: %v", err)
				b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Code: %v", res.ResponseMessage, res.ResponseCode))
			} else {
				b.Send(m.Sender, fmt.Sprintf("Successful Transaction\nResponse Message: %v\nBill Info: %v",
					res.ResponseMessage, res.BillInfo),
					&tb.SendOptions{
						ParseMode: "markdown",
					})
			}
		}
		// Count the Sum of the transactions.
		if errCounter > 0 {
			b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v\nAll Transaction(s) Completed with %d Errors\nMade with <3 By your friends at Solus!", time.Now().UTC(), errCounter))
			return
		}
		b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v, All Transaction(s) Completed Successfully\nMade with <3 By your friends at Solus!", time.Now().UTC()))

	})

	b.Handle("/payment", func(m *tb.Message) {
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

		expDate := p[2]
		pan := p[0]

		data := toStrings(p[3:])
		log.Printf("The len of data is: %v", len(data))

		if isOdd(len(p[3:])) {

			b.Send(m.Sender, "Wrong format. Please use biller_number amount pair")
			return
		}

		fields := dispatch(data)
		log.Printf("The fields output is: %#v", fields)
		var errCounter int

		for _, v := range fields {
			log.Printf("The bulked data is: %v, Type: %T", v[0], v)
			amountVal, _ := strconv.ParseFloat(v[1], 32)
			// amountVal := fmt.Sprintf("%.2f", fval)

			// generate ipin and generate uuid
			uuid := uuid.New().String()
			ipin, err := rsaEncrypt(key, p[1], uuid)

			if err != nil {
				log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
				b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
				return
			}
			biller, pre := getTopUp(v[0])
			res, err := billers(true, biller, pre+v[0], pan, ipin, expDate, uuid, float32(amountVal))
			if err != nil {
				errCounter++
				fmt.Printf("The error is: %v", err)
				b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Message: %v", res.ResponseMessage, res.ResponseCode))
			} else {
				b.Send(m.Sender, fmt.Sprintf("Successful Transaction\nResponse Message: %v\nBill Info: %v",
					res.ResponseMessage, res.BillInfo),
					&tb.SendOptions{
						ParseMode: "markdown",
					})
			}
		}
		// Count the Sum of the transactions.
		if errCounter > 0 {
			b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v\nAll Transaction(s) Completed with %d Errors\nMade with <3 By your friends at Solus!", time.Now().UTC(), errCounter))
			return
		}
		b.Send(m.Sender, fmt.Sprintf("Transaction Summary for %v, All Transaction(s) Completed Successfully\nMade with <3 By your friends at Solus!", time.Now().UTC()))

	})

	b.Handle("/invoices", func(m *tb.Message) {
		// get key

		if time.Since(m.Time()) > time.Minute {
			b.Send(m.Sender, "Message time out. Try again 🙏")
			return
		}

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

		expDate := p[2]
		pan := p[0]
		var errCounter int

		for _, v := range p[3:] {
			log.Printf("The bulked data is: %v, Type: %T", v[0], v)

			// generate ipin and generate uuid
			uuid := uuid.New().String()
			ipin, err := rsaEncrypt(key, p[1], uuid)

			if err != nil {
				log.Printf("The erorr is: %v. The IPIN is: %v\n", err, p[1])
				b.Send(m.Sender, "Failed to process the transaction. Code RSA_ERR")
				return
			}
			biller, pre := getBiller(v)
			res, err := billers(false, biller, pre+v, pan, ipin, expDate, uuid, 0)

			info := res.BillInfo
			if err != nil {
				errCounter++
				fmt.Printf("The error is: %v", err)
				b.Send(m.Sender, fmt.Sprintf("Transaction Failed.\nResponse Message: %v. \nResponse Code: %v", res.ResponseMessage, res.ResponseCode))
			} else {
				b.Send(m.Sender, fmt.Sprintf("Successful Transaction\nResponse Message: %v\nYour unpaid amount is: 💰%v💰\n\n\nBill Info: %v",
					res.ResponseMessage, info["billedAmount"], res.BillInfo),
					&tb.SendOptions{
						ParseMode: "markdown",
					})
			}
		}

		// Count the Sum of the transactions.
		if errCounter > 0 {
			b.Send(m.Sender, fmt.Sprintf("All Transaction(s) Completed with %d Errors\nMade with <3 By your friends at Solus!", errCounter))
			return
		}
		b.Send(m.Sender, fmt.Sprintf("All Transaction(s) Completed Successfully on :%v\nMade with <3 By your friends at Solus!", time.Now().UTC()))

	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		// all the text messages that weren't
		// captured by existing handlers
		b.Send(m.Sender, "command not found")
	})
	b.Start()
}

func getKey() (string, error) {
	k := Key{
		ApplicationId: "ACTSCon",
		TranDateTime:  "191124141930",
		UUID:          uuid.New().String(),
	}

	buf, _ := json.Marshal(&k)

	noebs, err := request(buf, ServerIP+"key")
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

	res, err := request(buf, ServerIP+"balance")
	if err != nil {
		return res, err
	}
	return res, nil
}

// billers(nec, payInfo, pan, ipin, expDate, uuid, amount)
func billers(isPayment bool, payeeId, personalInfo, pan, ipin, expDate, uuid string, amount float32) (Response, error) {

	/*
		zain top up: 0010010001
		mtn top up: 0010010003
		sudani top up: 0010010005
		nec top up: 0010020001
	*/

	// var pId string

	// switch p := payeeId; {
	// case p == zain:
	// 	pId = zain
	// case p == mtn:
	// 	pId = mtn
	// case p == sudani:
	// 	pId = sudani
	// case p == nec:
	// 	pId = nec
	// }

	k := Bills{
		Key: Key{
			ApplicationId: "ACTSCon",
			TranDateTime:  generateDate(),
			UUID:          uuid,
		},
		Card: Card{
			PAN:     pan,
			Expdate: expDate,
			IPIN:    ipin,
		},
		PayeeId:      payeeId,
		PersonalInfo: personalInfo,
		Amount: Amount{
			AmountNumber:     amount,
			TranCurrencyCode: "SDG",
		},
	}

	buf, _ := json.Marshal(&k)

	endpoint := "bill_payment"
	if !isPayment {
		endpoint = "bill_inquiry"
	}
	noebs, err := request(buf, ServerIP+endpoint)
	if err != nil {
		return noebs, err
	}
	return noebs, nil
}

func cardTransfer(toCard, pan, ipin, expDate, uuid string, amount float32) (Response, error) {

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

	noebs, err := request(buf, ServerIP+"p2p")
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

func tickerHandler() {
	for {
		select {
		case <-ticker.C:
			usdRate = rpcClient()
			c.Set("rate", usdRate, 24*time.Hour)
		}
	}
}
