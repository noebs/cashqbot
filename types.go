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
	ToCard string `json:"toCard`
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
	AmountNumber     float64 `json:"tranAmount"`
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
	ResponseMessage string                 `json:"responseMessage,omitempty"`
	ResponseStatus  string                 `json:"responseStatus,omitempty"`
	ResponseCode    int                    `json:"responseCode"`
	ReferenceNumber string                 `json:"referenceNumber,omitempty"`
	ApprovalCode    string                 `json:"approvalCode,omitempty"`
	Balance         map[string]interface{} `json:"balance,omitempty"`
	PaymentInfo     string                 `json:"paymentInfo,omitempty"`
	BillInfo        map[string]interface{} `json:"billInfo,omitempty"`
	Key             string                 `json:"pubKeyValue"`
}

type Error struct {
	Code    int
	Status  string
	Details Response
	Message string
}
