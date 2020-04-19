package main

// Key response
type Key struct {
	ApplicationId string `json:"applicationId" form:"applicationId" binding:"required"`
	TranDateTime  string `json:"tranDateTime" form:"tranDateTime" binding:"required"`
	UUID          string `json:"UUID" form:"UUID" binding:"required"`
}

type CardTransfer struct {
	Key
	Card
	ToCard string `json:"toCard"`
	Amount
}
type Bills struct {
	Key
	Card
	PayeeId      string `json:"payeeId"`
	PersonalInfo string `json:"paymentInfo"`
	Amount
}

type Amount struct {
	AmountNumber     float32 `json:"tranAmount"`
	TranCurrencyCode string  `json:"tranCurrencyCode"`
}

type Balance struct {
	Key
	Card
}

type Card struct {
	PAN     string `json:"PAN"`
	Expdate string `json:"expDate"`
	IPIN    string `json:"IPIN"`
}

type Noebs struct {
	Response `json:"ebs_response"`
}

type Response struct {
	ResponseMessage string `json:"responseMessage"`
	ResponseStatus  string `json:"responseStatus"`
	ResponseCode    int    `json:"responseCode"`
	// ReferenceNumber string                 `json:"referenceNumber"`
	// ApprovalCode    string                 `json:"approvalCode"`
	Balance     map[string]interface{} `json:"balance"`
	PaymentInfo string                 `json:"paymentInfo"`
	BillInfo    map[string]interface{} `json:"billInfo"`
	Key         string                 `json:"pubKeyValue"`
}

type Error struct {
	Code    int
	Status  string
	Details Response
	Message string
}

type necBill struct {
	SalesAmount  float64 `json:"SalesAmount"`
	FixedFee     float64 `json:"FixedFee"`
	Token        string  `json:"Token"`
	MeterNumber  string  `json:"MeterNumber"`
	CustomerName string  `json:"CustomerName"`
}
