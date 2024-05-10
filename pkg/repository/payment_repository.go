package repository

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"oosa/internal/helpers"

	linepay "github.com/chy168/line-pay-sdk-go"
	"github.com/gin-gonic/gin"
)

type PaymentRepository struct{}

type PaymentsRequest struct {
	Amount       int                         `json:"amount"`
	Currency     string                      `json:"currency"`
	OrderID      string                      `json:"orderId"`
	Packages     []PaymentsPackageRequest    `json:"packages"`
	RedirectUrls PaymentsRedirectUrlsRequest `json:"redirectUrls"`
	Options      PaymentsOptionsRequest      `json:"options"`
}

// `Id` required
// `Amount` required, valid amount `packages[].amount != sum(packages[].products[].quantity * packages[].products[].price)`
// `Name` required
type PaymentsPackageRequest struct {
	ID       string                          `json:"id"`
	Amount   int                             `json:"amount"`
	UserFee  int                             `json:"userFee,omitempty"`
	Name     string                          `json:"name"`
	Products []PaymentsPackageProductRequest `json:"products"`
}

// `Name` required
// `Quantity` required
// `Price` required
type PaymentsPackageProductRequest struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name"`
	ImageURL      string `json:"imageUrl,omitempty"`
	Quantity      int    `json:"quantity"`
	Price         int    `json:"price"`
	OriginalPrice int    `json:"originalPrice,omitempty"`
}

const (
	PaymentsConfirmUrlTypeClient string = "CLIENT"
	PaymentsConfirmUrlTypeServer string = "SERVER"
	PaymentsConfirmUrlTypeNone   string = "NONE"
)

type PaymentsRedirectUrlsRequest struct {
	AppPackageName string `json:"appPackageName,omitempty"`
	ConfirmURL     string `json:"confirmUrl"`
	ConfirmURLType string `json:"confirmUrlType,omitempty"`
	CancelURL      string `json:"cancelUrl"`
}

type PaymentsOptionsRequest struct {
	Payment       PaymentsOptionsPaymentRequest       `json:"payment"`
	Display       PaymentsOptionsDisplayRequest       `json:"display"`
	Shipping      PaymentsOptionsShippingRequest      `json:"shipping"`
	FamilyService PaymentsOptionsFamilyServiceRequest `json:"familyService"`
	Extra         PaymentsOptionsExtraRequest         `json:"extra"`
}

type PaymentsOptionsPaymentRequest struct {
	Capture bool   `json:"capture,omitempty"`
	PayType string `json:"payType,omitempty"` // NORMAL, PREAPPROVED
}

type PaymentsOptionsDisplayRequest struct {
	Locale                 string `json:"locale,omitempty"` // en, ja, ko, th, zh_TW, zh_CN
	CheckConfirmURLBrowser bool   `json:"checkConfirmUrlBrowser,omitempty"`
}

type PaymentsOptionsShippingRequest struct {
	ShippintType   string                                `json:"type,omitempty"`      // NO_SHIPPING, FIXED_ADDRESS, SHIPPING
	FeeAmount      string                                `json:"feeAmount,omitempty"` //why string?
	FeeInquiryURL  string                                `json:"feeInquiryUrl,omitempty"`
	FeeInquiryType string                                `json:"feeInquiryType,omitempty"` // CONDITION, FIXED
	Address        PaymentsOptionsShippingAddressRequest `json:"address"`
}

type PaymentsOptionsShippingAddressRequest struct {
	Country    string                                         `json:"country,omitempty"`
	PostalCode string                                         `json:"postalCode,omitempty"`
	State      string                                         `json:"state,omitempty"`
	City       string                                         `json:"city,omitempty"`
	Detail     string                                         `json:"detail,omitempty"`
	Optional   string                                         `json:"optional,omitempty"`
	Recipient  PaymentsOptionsShippingAddressRecipientRequest `json:"recipient,omitempty"`
}

type PaymentsOptionsShippingAddressRecipientRequest struct {
	FirstName         string `json:"firstName,omitempty"`
	LastName          string `json:"lastName,omitempty"`
	FirstNameOptional string `json:"firstNameOptional,omitempty"`
	LastNameOptional  string `json:"lastNameOptional,omitempty"`
	Email             string `json:"email,omitempty"`
	PhoneNo           string `json:"phoneNo,omitempty"`
}

type PaymentsOptionsFamilyServiceRequest struct {
	AddFriends []PaymentsOptionsFamilyServiceAddFriendsRequest `json:"addFriends"`
}

type PaymentsOptionsFamilyServiceAddFriendsRequest struct {
	AddType string   `json:"type,omitempty"` // line@
	IDs     []string `json:"ids,omitempty"`
}

type PaymentsOptionsExtraRequest struct {
	BranchName string `json:"branchName,omitempty"`
	BranchID   string `json:"branchId,omitempty"`
}

// response
type PaymentsResponse struct {
	ReturnCode    string               `json:"returnCode"`
	ReturnMessage string               `json:"returnMessage"`
	Info          PaymentsInfoResponse `json:"info"`
}

type PaymentsInfoResponse struct {
	TransactionID      int64                          `json:"transactionId"`
	PaymentAccessToken string                         `json:"paymentAccessToken"`
	PaymentURL         PaymentsInfoPaymentURLResponse `json:"paymentUrl"`
}

type PaymentsInfoPaymentURLResponse struct {
	Web string `json:"web"`
	App string `json:"app"`
}

// type PaymentRequest struct {
// 	Amount       int          `json:"amount"`
// 	Currency     string       `json:"currency"`
// 	OrderID      string       `json:"orderId"`
// 	Packages     []Package    `json:"packages"`
// 	RedirectUrls RedirectUrls `json:"redirectUrls"`
// }

// type PaymentsRequest struct {
// 	Amount       int                         `json:"amount"`
// 	Currency     string                      `json:"currency"`
// 	OrderID      string                      `json:"orderId"`
// 	Packages     []PaymentsPackageRequest    `json:"packages"`
// 	RedirectUrls PaymentsRedirectUrlsRequest `json:"redirectUrls"`
// 	Options      PaymentsOptionsRequest      `json:"options"`
// }

// type Package struct {
// 	ID       string    `json:"id"`
// 	Amount   int       `json:"amount"`
// 	UserFee  int       `json:"userFee"`
// 	Name     string    `json:"name"`
// 	Products []Product `json:"products"`
// }

// type Product struct {
// 	ID            string `json:"id"`
// 	Name          string `json:"name"`
// 	ImageURL      string `json:"imageUrl"`
// 	Quantity      int    `json:"quantity"`
// 	Price         int    `json:"price"`
// 	OriginalPrice int    `json:"originalPrice"`
// }

// type RedirectUrls struct {
// 	AppPackageName string `json:"appPackageName"`
// 	ConfirmUrl     string `json:"confirmUrl"`
// 	ConfirmUrlType string `json:"confirmUrlType"`
// 	CancelUrl      string `json:"cancelUrl"`
// }

const (
	baseURL       = "https://sandbox-api-pay.line.me"
	channelID     = "2004466370"
	channelSecret = "bb01e52f4d5d45bf324f0fa74455fd4b"
)

// MakeLinePayment handles Line payment request.
// @Summary MakeLinePayment
// @Description Make Line Payment
// @ID MakeLinePayment
// @Produce json
// @Tags Authentication
// @Success 200 {object} XXXXX
// @Failure 400 {object} XXXXX
// @Router /line/pay [get]
func (p PaymentRepository) MakeLinePayment(c *gin.Context) {
	client, err := linepay.NewClient(channelID, channelSecret, &linepay.Signer{ChannelId: channelID}, &linepay.ClientOpts{})
	if err != nil {
		helpers.ResponseError(c, fmt.Sprintf("New() error = %v", err.Error()))
		return
	}

	var requestBody linepay.PaymentsRequest
	if err := c.BindJSON(&requestBody); err != nil {
		helpers.ResponseBadRequestError(c, fmt.Sprintf("failed to parse request body: %s", err))
		return
	}

	res, err := client.PaymentsRequest(context.Background(), &requestBody)
	if err != nil {
		helpers.ResponseError(c, fmt.Sprintf("Test PaymentsRequest failed: %s", err.Error()))
	}

	// save transactionId and paymentUrl and products data to mongodb

	c.JSON(http.StatusOK, res)
}

// @Summary ConfirmLinePayment
// @Description Callback Confirm Line Payment
// @ID ConfirmLinePayment
// @Produce json
// @Tags Authentication
// @Success 200 {object} XXXXX
// @Failure 400 {object} XXXXX
// @Router /line/pay [get]
func (p PaymentRepository) ConfirmLinePayment(c *gin.Context) {
	fmt.Println("CALLBACK CONFIRM PAYMENT")

	// requestBody, err := ioutil.ReadAll(c.Request.Body)
	// if err != nil {
	// 	// Handle error, if any
	// 	fmt.Println("Error reading request body:", err)
	// 	return
	// }

	//minta https jadi belum bisa

	// fmt.Println("Request Body:", string(requestBody))

	// client, err := linepay.NewClient(channelID, channelSecret, &linepay.Signer{ChannelId: channelID}, &linepay.ClientOpts{})
	// if err != nil {
	// 	helpers.ResponseError(c, fmt.Sprintf("New() error = %v", err.Error()))
	// 	return
	// }

	// var requestBody linepay.PaymentsConfirmRequest
	// if err := c.BindJSON(&requestBody); err != nil {
	// 	helpers.ResponseBadRequestError(c, fmt.Sprintf("failed to parse request body: %s", err))
	// 	return
	// }

	// res, err := client.PaymentsConfirm(context.Background(), transactionId, &requestBody)
	// if err != nil {
	// 	helpers.ResponseError(c, fmt.Sprintf("Test PaymentsRequest failed: %s", err.Error()))
	// }
	// fmt.Println(res)
	// fmt.Println(requestBody)

	// c.JSON(http.StatusOK, requestBody)
}

func getAuthHeader(secret, uri, body, nonce string) string {
	data := secret + uri + body + nonce
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
