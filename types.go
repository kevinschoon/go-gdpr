package gdpr

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const ApiVersion = "0.1"

// SubjectType is the type of request
// that is being made.
type SubjectType string

func (s SubjectType) Valid() bool {
	_, ok := SubjectTypeMap[string(s)]
	return ok
}

func (s *SubjectType) UnmarshalJSON(raw []byte) error {
	str := strings.Replace(string(raw), "\"", "", -1)
	if _, ok := SubjectTypeMap[str]; !ok {
		return fmt.Errorf("bad subject type: %s", str)
	}
	*s = SubjectType(str)
	return nil
}

const (
	SUBJECT_ACCESS      = SubjectType("access")
	SUBJECT_PORTABILITY = SubjectType("portability")
	SUBJECT_ERASURE     = SubjectType("erasure")
)

var SubjectTypeMap = map[string]SubjectType{
	"access":      SUBJECT_ACCESS,
	"portability": SUBJECT_PORTABILITY,
	"erasure":     SUBJECT_ERASURE,
}

type IdentityType string

const (
	IDENTITY_CONTROLLER_CUSTOMER_ID   = IdentityType("controller_customer_id")
	IDENTITY_ANDROID_ADVERTISING_ID   = IdentityType("android_advertising_id")
	IDENTITY_ANDROID_ID               = IdentityType("android_id")
	IDENTITY_EMAIL                    = IdentityType("email")
	IDENTITY_FIRE_ADVERTISING_ID      = IdentityType("fire_advertising_id")
	IDENTITY_IOS_ADVERTISING_ID       = IdentityType("ios_advertising_id")
	IDENTITY_IOS_VENDOR_ID            = IdentityType("ios_vendor_id")
	IDENTITY_MICROSOFT_ADVERTISING_ID = IdentityType("microsoft_advertising_id")
	IDENTITY_MICROSOFT_PUBLISHER_ID   = IdentityType("microsoft_publisher_id")
	IDENTITY_ROKU_PUBLISHER_ID        = IdentityType("roku_publisher_id")
	IDENTITY_ROKU_ADVERTISING_ID      = IdentityType("roku_advertising_id")
)

func (i IdentityType) Valid() bool {
	_, ok := IdentityTypeMap[string(i)]
	return ok
}

func (i *IdentityType) UnmarshalJSON(raw []byte) error {
	str := strings.Replace(string(raw), "\"", "", -1)
	if _, ok := IdentityTypeMap[str]; !ok {
		return fmt.Errorf("bad identity type: %s", str)
	}
	*i = IdentityType(str)
	return nil
}

var IdentityTypeMap = map[string]IdentityType{
	"controller_customer_id":   IDENTITY_CONTROLLER_CUSTOMER_ID,
	"android_advertising_id":   IDENTITY_ANDROID_ADVERTISING_ID,
	"android_id":               IDENTITY_ANDROID_ID,
	"email":                    IDENTITY_EMAIL,
	"fire_advertising_id":      IDENTITY_FIRE_ADVERTISING_ID,
	"ios_advertising_id":       IDENTITY_IOS_ADVERTISING_ID,
	"ios_vendor_id":            IDENTITY_IOS_VENDOR_ID,
	"microsoft_advertising_id": IDENTITY_MICROSOFT_ADVERTISING_ID,
	"microsoft_publisher_id":   IDENTITY_MICROSOFT_PUBLISHER_ID,
	"roku_publisher_id":        IDENTITY_ROKU_PUBLISHER_ID,
	"roku_advertising_id":      IDENTITY_ROKU_ADVERTISING_ID,
}

type IdentityFormat string

const (
	FORMAT_RAW    = IdentityFormat("raw")
	FORMAT_SHA1   = IdentityFormat("sha1")
	FORMAT_MD5    = IdentityFormat("md5")
	FORMAT_SHA256 = IdentityFormat("sha256")
)

var IdentityFormatMap = map[string]IdentityFormat{
	"raw":    FORMAT_RAW,
	"sha1":   FORMAT_SHA1,
	"md5":    FORMAT_MD5,
	"sha256": FORMAT_SHA256,
}

func (i IdentityFormat) Valid() bool {
	_, ok := IdentityFormatMap[string(i)]
	return ok
}

func (i *IdentityFormat) UnmarshalJSON(raw []byte) error {
	str := strings.Replace(string(raw), "\"", "", -1)
	if _, ok := IdentityFormatMap[str]; !ok {
		return fmt.Errorf("bad identity format: %s", str)
	}
	*i = IdentityFormat(str)
	return nil
}

// RequestStatus represents the status of a GDPR request
type RequestStatus string

const (
	STATUS_PENDING     = RequestStatus("pending")
	STATUS_IN_PROGRESS = RequestStatus("in_progress")
	STATUS_COMPLETED   = RequestStatus("completed")
	STATUS_CANCELLED   = RequestStatus("cancelled")
)

var RequestStatusMap = map[string]RequestStatus{
	"pending":     STATUS_PENDING,
	"in_progress": STATUS_IN_PROGRESS,
	"completed":   STATUS_COMPLETED,
	"cancelled":   STATUS_CANCELLED,
}

func (r RequestStatus) Valid() bool {
	_, ok := RequestStatusMap[string(r)]
	return ok
}

func (r *RequestStatus) UnmarshalJSON(raw []byte) error {
	str := strings.Replace(string(raw), "\"", "", -1)
	if _, ok := RequestStatusMap[str]; !ok {
		return fmt.Errorf("bad request status format: %s", str)
	}
	*r = RequestStatus(str)
	return nil
}

type Identity struct {
	Type   IdentityType   `json:"identity_type"`
	Format IdentityFormat `json:"identity_format"`
	Value  string         `json:"identity_value,omitempty"`
}

// Request represents a GDPR request
type Request struct {
	SubjectRequestId   string      `json:"subject_request_id"`
	SubjectRequestType SubjectType `json:"subject_request_type"`
	SubmittedTime      time.Time   `json:"submitted_time"`
	ApiVersion         string      `json:"api_version"`
	StatusCallbackUrls []string    `json:"status_callback_urls"`
	SubjectIdentities  []Identity  `json:"subject_identities"`
	// TODO
	Extensions json.RawMessage `json:"extensions"`
}

func (r Request) Base64() string {
	raw, _ := json.Marshal(r)
	return base64.StdEncoding.EncodeToString(raw)
}

func (r Request) Signature() string {
	hash := sha256.New()
	json.NewEncoder(hash).Encode(r)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

type Response struct {
	ControllerId           string    `json:"controller_id"`
	ExpectedCompletionTime time.Time `json:"expected_completion_time"`
	ReceivedTime           time.Time `json:"received_time"`
	EncodedRequest         string    `json:"encoded_request"`
	SubjectRequestId       string    `json:"subject_request_id"`
}

type DiscoveryResponse struct {
	ApiVersion                   string        `json:"api_version"`
	SupportedIdentities          []Identity    `json:"supported_identities"`
	SupportedSubjectRequestTypes []SubjectType `json:"supported_subject_request_types"`
	ProcessorCertificate         string        `json:"processor_certificate"`
}

type StatusResponse struct {
	ControllerId           string        `json:"controller_id"`
	ExpectedCompletionTime time.Time     `json:"expected_completion_time"`
	SubjectRequestId       string        `json:"subject_request_id"`
	RequestStatus          RequestStatus `json:"request_status"`
	ApiVersion             string        `json:"api_version"`
	ResultsUrl             string        `json:"results_url"`
}

type errMsg struct {
	Error errMsgInner `json:"error"`
}

type errMsgInner struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Errors  []Error `json:"errors"`
}

type ErrorResponse struct {
	Code    int     `json:"-"`
	Message string  `json:"-"`
	Errors  []Error `json:"-"`
}

func (e *ErrorResponse) UnmarshalJSON(raw []byte) error {
	msg := &errMsg{
		Error: errMsgInner{},
	}
	err := json.Unmarshal(raw, &msg)
	if err != nil {
		return err
	}
	e.Code = msg.Error.Code
	e.Message = msg.Error.Message
	e.Errors = msg.Error.Errors
	return nil
}

func (e ErrorResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		errMsg{
			Error: errMsgInner{
				Message: e.Message,
				Code:    e.Code,
				Errors:  e.Errors,
			},
		},
	)
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("code=%d,message=%s,nested_errors=%d", e.Code, e.Message, len(e.Errors))
}

type Error struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("domain=%s,reason=%s,message=%s", e.Domain, e.Reason, e.Message)
}

type CallbackRequest struct {
	ControllerId           string        `json:"controller_id"`
	ExpectedCompletionTime time.Time     `json:"expected_completion_time"`
	StatusCallbackUrl      string        `json:"status_callback_url"`
	SubjectRequestId       string        `json:"subject_request_id"`
	RequestStatus          RequestStatus `json:"request_status"`
	ResultsUrl             string        `json:"results_url"`
}

type CancellationResponse struct {
	ControllerId     string    `json:"controller_id"`
	SubjectRequestId string    `json:"subject_request_id"`
	ReceivedTime     time.Time `json:"ReceivedTime"`
	EncodedRequest   string    `json:"encoded_request"`
	ApiVersion       string    `json:"api_version"`
}
