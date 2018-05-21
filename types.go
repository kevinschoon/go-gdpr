package gdpr

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

const ApiVersion = "0.1"

type SubjectType string

func (s SubjectType) Valid() bool {
	_, ok := SubjectTypeMap[string(s)]
	return ok
}

func (s *SubjectType) UnmarshalJSON(raw []byte) error {
	if _, ok := SubjectTypeMap[string(raw)]; !ok {
		return fmt.Errorf("bad subject type: %s", string(raw))
	}
	*s = SubjectType(string(raw))
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
	if _, ok := IdentityTypeMap[string(raw)]; !ok {
		return fmt.Errorf("bad identity type: %s", string(raw))
	}
	*i = IdentityType(string(raw))
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
	if _, ok := IdentityFormatMap[string(raw)]; !ok {
		return fmt.Errorf("bad identity format: %s", string(raw))
	}
	*i = IdentityFormat(string(raw))
	return nil
}

type Identity struct {
	Type   IdentityType   `json:"identity_type"`
	Format IdentityFormat `json:"identity_format"`
	Value  string         `json:"identity_value"`
}

/*
POST /opengdpr_requests HTTP/1.1
Host: example-processor.com
Accept: application/json
Content Type: application/json
{
  "subject_request_id": "a7551968-d5d6-44b2-9831-815ac9017798",
  "subject_request_type": "erasure",
  "submitted_time": "2018-10-02T15:00:00Z",
  "subject_identities": [
    {
      "identity_type": "email",
      "identity_value": "johndoe@example.com",
      "identity_format": "raw"
    }
  ],
  "api_version": "0.1",
  "status_callback_urls": [
    "https://examplecontroller.com/opengdpr_callbacks"
  ],
  "extensions": {
    "example-processor.com": {
      "foo-processor-custom-id":123456,
      "property_id": "123456",
    },
    "example-other-processor.com": {
      "foo-other-processor-custom-id":654321
    }
  }
}
*/
type Request struct {
	SubjectRequestId   string    `json:"subject_request_id"`
	SubjectRequestType string    `json:"subject_request_type"`
	SubmittedTime      time.Time `json:"submitted_time"`
	ApiVersion         string    `json:"api_version"`
	StatusCallbackUrls []string  `json:"status_callback_urls"`
	// Extensions TODO TODO
}

func (r Request) Base64() string {
	raw, _ := json.Marshal(r)
	return base64.StdEncoding.EncodeToString(raw)
}

/*
HTTP/1.1 201 Created
Content-Type: application/json
X-OpenGDPR-Processor-Domain: example-processor.com
X-OpenGDPR-Signature:
kiGlog3PdQx+FQmB8wYwFC1fekbJG7Dm9WdqgmXc9uKkFRSM4uPzylLi7j083461xLZ+mUloo3tpsmyI
Zpt5eMfgo7ejXPh6lqB4ZgCnN6+1b6Q3NoNcn/+11UOrvmDj772wvg6uIAFzsSVSjMQxRs8LAmHqFO4c
F2pbuoPuK2diHOixxLj6+t97q0nZM7u3wmgkwF9EHIo3C6G1SI04/odvyY/VdMZgj3H1fLnz+X5rc42/
wU4974u3iBrKgUnv0fcB4YB+L6Q3GsMbmYzuAbe0HpVA17ud/bVoyQZAkrW2yoSy1x4Ts6XKba6pLifI
Hf446Bubsf5r7x1kg6Eo7B8zur666NyWOYrglkOzU4IYO8ifJFRZZXazOgk7ggn9obEd78GBc3kjKKZd
waCrLx7WV5y9TMDCf+2FILOJM/MwTUy1dLZiaFHhGdzld2AjbjK1CfVzyPssch0iQYYtbR49GhumvkYl
11S4oDfu0c3t/xUCZWg0hoR3XL3B7NjcrlrQinB1KbyTNZccKR0F4Lk9fDgwTVkrAg152UqPyzXxpdzX
jfkDkSEgAevXQwVJWBNf18bMIEgdH2usF/XauQoyrne7rcMIWBISPgtBPj3mhcrwscjGVsxqJva8KCVC
KD/4Axmo9DISib5/7A6uczJxQG2Bcrdj++vQqK2succ=
{
    "controller_id":"example_controller_id",
    "expected_completion_time":"2018-11-01T15:00:01Z",
    "received_time":"2018 10 02T15:00:01Z",
    "encoded_request":"<BASE64 ENCODED REQUEST>",
    "subject_request_id":"a7551968-d5d6-44b2-9831-815ac9017798"
}
*/
type Response struct {
	ControllerId           string    `json:"controller_id"`
	ExpectedCompletionTime time.Time `json:"expected_completion_time"`
	ReceivedTime           time.Time `json:"received_time"`
	EncodedRequest         string    `json:"encoded_request"`
	SubjectRequestId       string    `json:"subject_request_id"`
}

/*
HTTP/1.1 200 OK
Content Type: application/json
{
   "api_version":"0.1",
   "supported_identities":[
      {
         "identity_type":"email",
         "identity_format":"raw"
      },
      {
         "identity_type":"email",
         "identity_format":"sha256"
      }
   ],
   "supported_subject_request_types":[
      "erasure"
   ],
   "processor_certificate":"https://exampleprocessor.com/cert.pem"
}
*/
type DiscoveryResponse struct {
	ApiVersion                   string        `json:"api_version"`
	SupportedIdentities          []Identity    `json:"supported_identities"`
	SupportedSubjectRequestTypes []SubjectType `json:"supported_subject_request_types"`
	ProcessorCertificate         string        `json:"processor_certificate"`
}

/*
HTTP/1.1 200 OK
Content Type: application/json
X-OpenGDPR-Processor-Domain: example-processor.com
X-OpenGDPR-Signature:
kiGlog3PdQx+FQmB8wYwFC1fekbJG7Dm9WdqgmXc9uKkFRSM4uPzylLi7j083461xLZ+mUloo3tpsmyI
Zpt5eMfgo7ejXPh6lqB4ZgCnN6+1b6Q3NoNcn/+11UOrvmDj772wvg6uIAFzsSVSjMQxRs8LAmHqFO4c
F2pbuoPuK2diHOixxLj6+t97q0nZM7u3wmgkwF9EHIo3C6G1SI04/odvyY/VdMZgj3H1fLnz+X5rc42/
wU4974u3iBrKgUnv0fcB4YB+L6Q3GsMbmYzuAbe0HpVA17ud/bVoyQZAkrW2yoSy1x4Ts6XKba6pLifI
Hf446Bubsf5r7x1kg6Eo7B8zur666NyWOYrglkOzU4IYO8ifJFRZZXazOgk7ggn9obEd78GBc3kjKKZd
waCrLx7WV5y9TMDCf+2FILOJM/MwTUy1dLZiaFHhGdzld2AjbjK1CfVzyPssch0iQYYtbR49GhumvkYl
11S4oDfu0c3t/xUCZWg0hoR3XL3B7NjcrlrQinB1KbyTNZccKR0F4Lk9fDgwTVkrAg152UqPyzXxpdzX
jfkDkSEgAevXQwVJWBNf18bMIEgdH2usF/XauQoyrne7rcMIWBISPgtBPj3mhcrwscjGVsxqJva8KCVC
KD/4Axmo9DISib5/7A6uczJxQG2Bcrdj++vQqK2succ=
{
    "controller_id":"example_controller_id",
    "expected_completion_time":"2018-11-01T15:00:01Z",
    "subject_request_id":"a7551968-d5d6-44b2-9831-815ac9017798",
    "request_status":"pending",
    "api_version":"0.1",
    "results_url":"https://exampleprocessor.com/secure/d188d4ba-12db-48a0-898c-cd0f8ba7b345"
}
*/
type StatusResponse struct {
	ControllerId           string    `json:"controller_id"`
	ExpectedCompletionTime time.Time `json:"expected_completion_time"`
	SubjectRequestId       string    `json:"subject_request_id"`
	RequestStatus          string    `json:"request_status"`
	ApiVersion             string    `json:"api_version"`
	ResultsUrl             string    `json:"results_url"`
}

/*
HTTP/1.1 400 Bad Request
Content Type: application/json;charset=UTF-8
Cache Control: no store
Pragma: no cache
{
  "error": {
    "code": 400,
    "message": "subject_request_id field is required",
    "errors": [
      {
        "domain": "Validation",
        "reason": "IllegalArgumentException",
        "message": "subject_request_id field is required."
      }
    ]
  }
}
*/
type ErrorResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Errors  []Error `json:"errors"`
}

type Error struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (e ErrorResponse) Error() string { return e.Message }

/*
POST /opengdpr_callbacks HTTP/1.1
Host: examplecontroller.com
Content Type: application/json
X-OpenGDPR-Processor-Domain: example-processor.com
X-OpenGDPR-Signature:
kiGlog3PdQx+FQmB8wYwFC1fekbJG7Dm9WdqgmXc9uKkFRSM4uPzylLi7j083461xLZ+mUloo3tpsmyI
Zpt5eMfgo7ejXPh6lqB4ZgCnN6+1b6Q3NoNcn/+11UOrvmDj772wvg6uIAFzsSVSjMQxRs8LAmHqFO4c
F2pbuoPuK2diHOixxLj6+t97q0nZM7u3wmgkwF9EHIo3C6G1SI04/odvyY/VdMZgj3H1fLnz+X5rc42/
wU4974u3iBrKgUnv0fcB4YB+L6Q3GsMbmYzuAbe0HpVA17ud/bVoyQZAkrW2yoSy1x4Ts6XKba6pLifI
Hf446Bubsf5r7x1kg6Eo7B8zur666NyWOYrglkOzU4IYO8ifJFRZZXazOgk7ggn9obEd78GBc3kjKKZd
waCrLx7WV5y9TMDCf+2FILOJM/MwTUy1dLZiaFHhGdzld2AjbjK1CfVzyPssch0iQYYtbR49GhumvkYl
11S4oDfu0c3t/xUCZWg0hoR3XL3B7NjcrlrQinB1KbyTNZccKR0F4Lk9fDgwTVkrAg152UqPyzXxpdzX
jfkDkSEgAevXQwVJWBNf18bMIEgdH2usF/XauQoyrne7rcMIWBISPgtBPj3mhcrwscjGVsxqJva8KCVC
KD/4Axmo9DISib5/7A6uczJxQG2Bcrdj++vQqK2succ=
{
    "controller_id":"example_controller_id",
    "expected_completion_time":"2018-11-01T15:00:01Z",
    "status_callback_url":"https://examplecontroller.com/opengdpr_callbacks",
    "subject_request_id":"a7551968-d5d6-44b2-9831-815ac9017798",
    "request_status":"pending",
    "results_url":"https://exampleprocessor.com/secure/d188d4ba-12db-48a0-898c-cd0f8ba7b345"
}
*/

type CallbackRequest struct {
	ControllerId           string    `json:"controller_id"`
	ExpectedCompletionTime time.Time `json:"expected_completion_time"`
	StatusCallbackUrl      string    `json:"status_callback_url"`
	SubjectRequestId       string    `json:"subject_request_id"`
	RequestStatus          string    `json:"request_status"`
	ResultsUrl             string    `json:"results_url"`
}

/*
HTTP/1.1 202 Accepted
Content Type: application/json
X-OpenGDPR-Processor-Domain: example-processor.com
X-OpenGDPR-Signature:
kiGlog3PdQx+FQmB8wYwFC1fekbJG7Dm9WdqgmXc9uKkFRSM4uPzylLi7j083461xLZ+mUloo3tpsmyI
Zpt5eMfgo7ejXPh6lqB4ZgCnN6+1b6Q3NoNcn/+11UOrvmDj772wvg6uIAFzsSVSjMQxRs8LAmHqFO4c
F2pbuoPuK2diHOixxLj6+t97q0nZM7u3wmgkwF9EHIo3C6G1SI04/odvyY/VdMZgj3H1fLnz+X5rc42/
wU4974u3iBrKgUnv0fcB4YB+L6Q3GsMbmYzuAbe0HpVA17ud/bVoyQZAkrW2yoSy1x4Ts6XKba6pLifI
Hf446Bubsf5r7x1kg6Eo7B8zur666NyWOYrglkOzU4IYO8ifJFRZZXazOgk7ggn9obEd78GBc3kjKKZd
waCrLx7WV5y9TMDCf+2FILOJM/MwTUy1dLZiaFHhGdzld2AjbjK1CfVzyPssch0iQYYtbR49GhumvkYl
11S4oDfu0c3t/xUCZWg0hoR3XL3B7NjcrlrQinB1KbyTNZccKR0F4Lk9fDgwTVkrAg152UqPyzXxpdzX
jfkDkSEgAevXQwVJWBNf18bMIEgdH2usF/XauQoyrne7rcMIWBISPgtBPj3mhcrwscjGVsxqJva8KCVC
KD/4Axmo9DISib5/7A6uczJxQG2Bcrdj++vQqK2succ=
{
  "controller_id": "example_controller_id",
  "subject_request_id": "a7551968-d5d6-44b2-9831-815ac9017798",
  "received_time": "2018-10-02T15:00:01Z",
  "encoded_request": "<BASE64 ENCODED REQUEST>",
  "api_version": "0.1"
}
*/
type CancellationResponse struct {
	ControllerId     string    `json:"controller_id"`
	SubjectRequestId string    `json:"subject_request_id"`
	ReceivedTime     time.Time `json:"ReceivedTime"`
	EncodedRequest   string    `json:"encoded_request"`
	ApiVersion       string    `json:"api_version"`
}
