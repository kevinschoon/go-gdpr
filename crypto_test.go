package gdpr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Keys generated via:
// openssl req -x509 -newkey rsa:1024 -keyout key.pem -out cert.pem -days 365 -nodes

var keyPairOne = [][]byte{
	[]byte(`
-----BEGIN PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAMIHALmy8o85B5Th
QB9fuooQESfiGvzCWyCe3iWpFEio50zKcihUdiFmuwY/qMnInxUhohBmkQMC6hpT
ikcv6fzv+2dYWo9T0meijT33DVcOkwLnaNMccVcK/654bsFDxGg0FQk1srydGDGk
7Xhf0Vif1+YqcCpxNeuSyzkzwerFAgMBAAECgYBRk2QoryXwNYgMfk/ZYQQqu+qa
nCPAlW5+3oyDxPy0N99Xl947OpeYH3sOe4FZpTHNTqC2yIi7fWQzwV/n4is74jBQ
EHlz2CubgSMxIiofLdeDk01oSgaiDEcBxEcoyCFqetip8+oFwqz5rsMuwjQIkjJw
nJyyG5Vzkf1bPaP+WQJBAP1NNps/x7rzP3nqlLIBFYjTm1z8iMUCeohpBgTawh7v
e0XPlSgQO+bpLYCAILM6DjxgKiMQDeXgGCqlse2Rh6cCQQDEGCQPBGeDxJtNdTU2
5UJSYaWKTWlSjltNmjApBQ5xdfemYv2lNLTh78xFAuk2MxwGa3aKxeU0lNgc+ZvC
sYezAkAd9S7bJ6zwoGpGegcCEnzAhP5f/gITAtsJHRq4IkNJM1uqAwYKCfl7suJN
y1mSuPAMFfeF1BVAtcNF7/jeNxMLAkAoI6Tl6gniYBFGJrLQ3NbZlCFVkQj5HCi2
VtR64Q0WzoX16hdvhL1t7i8LBVCFhqq66a5nM6D6RWmDbNikXsCfAkEAjAnFCbHM
8eCBQDM8weRXTfJdwEIpsYQFBVf9ZZ7fcrJwjcrUUl62uWtuwibF33kLKuMiiZWL
BY4y1QFbJOg+Tw==
-----END PRIVATE KEY-----
`),
	[]byte(`
-----BEGIN CERTIFICATE-----
MIICWzCCAcSgAwIBAgIJAKim3pmMrx+WMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTgwNTI4MTM0MTM3WhcNMTkwNTI4MTM0MTM3WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKB
gQDCBwC5svKPOQeU4UAfX7qKEBEn4hr8wlsgnt4lqRRIqOdMynIoVHYhZrsGP6jJ
yJ8VIaIQZpEDAuoaU4pHL+n87/tnWFqPU9Jnoo099w1XDpMC52jTHHFXCv+ueG7B
Q8RoNBUJNbK8nRgxpO14X9FYn9fmKnAqcTXrkss5M8HqxQIDAQABo1MwUTAdBgNV
HQ4EFgQU/0OQ9gNZKIa07UjEp6sCC8rMl9MwHwYDVR0jBBgwFoAU/0OQ9gNZKIa0
7UjEp6sCC8rMl9MwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBs
Gvrf0qeHXvXwSaFt6g8K7ZKooZv3lPuy7UUVXdUbmHAMy7/TQtgVKw0wYkr3hW+G
x4FBCkNUI1zSOUQwZh3E4AgBF0qof00SJnJi3mAe6jrSjcrGEfnZtuypzJ00MzfY
bmegdr80vALqvPyrqdalB6NutZ1A6d/67r/b0wvppw==
-----END CERTIFICATE-----
`)}

var keyPairTwo = [][]byte{
	[]byte(`
-----BEGIN PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBAKLTNyd+kpOea58s
x14hK13/9TAq4fBpeN/LYAUcF8b//tjYO1qF4hH17JVAwxsz697CED8OJHZJIJgv
H/LmS9n8cSq0gJkS1fyTsJFgtXhbEEJ9v6SE38YSqaib3JjL9mDTEIHecrrQbIK+
mL+IzEOBm/LNlOzIr25rqWmlwaRBAgMBAAECgYAI/IFBw8GRNiAYc984+bmsAXFl
zCgWHawXJeFRxuAlEoHdM+nqsBLvDNSW1DEwciglbi55XG10vcp3u7oWrNEoyzPJ
v6God0xhyK40bnW+9+eMtD8YdA2iUlvFKoRcDRzbjq4FuDauEQGcI8g4hIY4Xcsm
lTkZa+6pTRRMcjWeVQJBANF5SLFeOzmTbOE/OECaUzd6I0v/C1MCEQrmE7dAXe2E
gbG5A1HNwQA2WE1UjipdLViZjqn/lNfkuJKJ79yQuQsCQQDG/XpLXmDlKUnD2sw7
e/ACo3Si8FLC4bcDOk7f1L2PdtluF/qs4/KLptUOyks4cf8D4rotX4Hot6ej7VQY
vV9jAkBunPv85USi962EGC0tOBD/d5iR9eDV+X5kYfBBUVUIKnOOFKOjG+JxqUDh
vOfBiSh748KJFHRVuOqaPwqRTz7XAkA3MaO0OA9kQNmHC69OaIggEzqM31/1Uioz
KP8rspSJsIuKr/gF8IwcFEBQg+ftViFH8KF3aGBeLmK/Y1rKKezFAkEAuE9F6YxB
/jr7TxMUj9zap6xBeoG80yZxJRSyFTOIWJNNVl4yXMNs1C70xwhfcmQy7ZPUiTfp
RgP7AjghsiGXkA==
-----END PRIVATE KEY-----
`),
	[]byte(`
-----BEGIN CERTIFICATE-----
MIICWzCCAcSgAwIBAgIJAOdhyN22QoyFMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTgwNTI4MTQ1ODMyWhcNMTkwNTI4MTQ1ODMyWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKB
gQCi0zcnfpKTnmufLMdeIStd//UwKuHwaXjfy2AFHBfG//7Y2DtaheIR9eyVQMMb
M+vewhA/DiR2SSCYLx/y5kvZ/HEqtICZEtX8k7CRYLV4WxBCfb+khN/GEqmom9yY
y/Zg0xCB3nK60GyCvpi/iMxDgZvyzZTsyK9ua6lppcGkQQIDAQABo1MwUTAdBgNV
HQ4EFgQUlKjtdFV/klc8ZxeOtSpzu8fstKQwHwYDVR0jBBgwFoAUlKjtdFV/klc8
ZxeOtSpzu8fstKQwDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQAw
1W3mZP332f2Ay1YpK4zEKQzozfkhXj7sHQtOoEoWDLg41C9vYnEJDysBGD9QYiYj
GFPq3aNGZDqWBSn07OLvq4uyWUuutiuovnyTjVK7/in6rOGWLpAhTRr/BAlJgiQ2
vDiZORDvtNXvX6YoawjXH2NiDY1/STvw82w4h7etqw==
-----END CERTIFICATE-----
`)}

func TestRsaSigner(t *testing.T) {
	expectedSig := "Zcfr4BKJKKsDN/l8rTl44p4R3Mx9hXV5atssYGwU44naAgnP55B2miwPLyRDErWsrxs84wdsl6Uu00xBh1pfGiqjb9EfGBLdj9I1Pm2HC/r9L4pzx/wZWiqNTRQ0Dg+WJ2H4cEIP9f8+kdMa6Jek7+ks8hc1CwZhuzxBWkl4E2E="
	resp := &Response{
		ControllerId:     "controller-1234",
		SubjectRequestId: "request-1234",
	}
	respRaw, _ := json.Marshal(resp)
	signer, err := NewSigner(&KeyOptions{KeyBytes: keyPairOne[0]})
	assert.NoError(t, err)
	sig, err := signer.Sign(respRaw)
	assert.NoError(t, err)
	assert.Equal(t, expectedSig, sig)
	// Change the underlying key
	signer, err = NewSigner(&KeyOptions{KeyBytes: keyPairTwo[0]})
	assert.NoError(t, err)
	sig, err = signer.Sign(respRaw)
	assert.NoError(t, err)
	assert.NotEqual(t, sig, expectedSig)
}

func TestRsaVerifier(t *testing.T) {
	expectedSig := "Zcfr4BKJKKsDN/l8rTl44p4R3Mx9hXV5atssYGwU44naAgnP55B2miwPLyRDErWsrxs84wdsl6Uu00xBh1pfGiqjb9EfGBLdj9I1Pm2HC/r9L4pzx/wZWiqNTRQ0Dg+WJ2H4cEIP9f8+kdMa6Jek7+ks8hc1CwZhuzxBWkl4E2E="
	resp := &Response{
		ControllerId:     "controller-1234",
		SubjectRequestId: "request-1234",
	}
	respRaw, _ := json.Marshal(resp)
	verifier, err := NewVerifier(&KeyOptions{KeyBytes: keyPairOne[1]})
	assert.NoError(t, err)
	err = verifier.Verify(respRaw, expectedSig)
	assert.NoError(t, err)
	verifier, err = NewVerifier(&KeyOptions{KeyBytes: keyPairTwo[1]})
	assert.NoError(t, err)
	err = verifier.Verify(respRaw, expectedSig)
	assert.Error(t, err)
}

func TestSignerVerifier(t *testing.T) {
	signer, err := NewSigner(&KeyOptions{KeyBytes: keyPairOne[0]})
	assert.NoError(t, err)
	resp := &Response{
		ControllerId:     "controller-1234",
		SubjectRequestId: "request-1234",
	}
	respRaw, _ := json.Marshal(resp)
	sig, err := signer.Sign(respRaw)
	assert.NoError(t, err)
	verifier, err := NewVerifier(&KeyOptions{KeyBytes: keyPairOne[1]})
	assert.NoError(t, err)
	err = verifier.Verify(respRaw, sig)
	assert.NoError(t, err)
}

func BenchmarkSignVerify(b *testing.B) {
	resp := &Response{
		ControllerId:     "controller-1234",
		SubjectRequestId: "request-1234",
	}
	respRaw, _ := json.Marshal(resp)
	signer, _ := NewSigner(&KeyOptions{KeyBytes: keyPairOne[0]})
	verifier, _ := NewVerifier(&KeyOptions{KeyBytes: keyPairOne[1]})
	for n := 0; n < b.N; n++ {
		sig, _ := signer.Sign(respRaw)
		verifier.Verify(respRaw, sig)
	}
}
