package caleta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// testing keys from Caleta signature sample project https://bitbucket.org/paulocaleta/signaturetest/src/master/keys/
const testingPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAwUlXX2eDngAIDlH8XX1N4kCff98Ev2ulCXMpu1wvR7LrmZ5G
QpPtt5VCr7mcy0C8tUe8UZli/jYbicdAYGAGb8t4d3TqDpewx9igSvWTLMwvZi7+
f0fudpdUA3cFZOpKiWdOQSlbIKuf3IAgP/g1ETb46mOjSSsd7b8gL17m3xUZMnza
vGujM4H6t3tTKLnn2CFLmlxXyk/El/FQohQps6ihr+MBQxwwnVY9BT0vPxXf2pw5
hoHgzqsayT5IDGkqCB3h0Ng6QWumNfzqiJZDeodNhuW80MZjCThgqC1F2yZukvcM
BWHLYnmnYIFZaOPEFuzW9OLJZMdc+XCXmixMRwIDAQABAoIBAQCGpKV1szEvXkw+
VYRcR856XpP9SokPI1hbBds7RhM1egC/CU9eT5gX+6TxrnX37IfILEmV+ijIdz6l
sqQ4voudUvr/E/V75tVh0ZiPvxQf77jREMl+Nsh52h7PLxkV3FPB9bHAmKN/Va7N
tn9AsJGfBVFOTcxQSvXVSP+NoClpmh3VDjZQKCjEON+0p0gjsqTKYG1EzASTAAP3
pU3dj/HwNMptzJZFblu6K5m4DQxfm5ZYyfAXO+2dl0D59GzgM5Bf4fGwbGPZT7nv
rOSRklqS66VG5kHZMUW775pvEnBIqa8xfPkqL8azHysK/3mMmhASkC7u9LLfiPs1
tBnoSz4JAoGBAOYbdLvMNDHxE5irkBEZRlt+5HOq3VE/CpBjsV5MVLicK3ssTl0X
4j2REL2xEONaAPlUN+f7XvDlKQcxfelFp8Z2GI49bJLLMTME6GeDoS0CXw/90OQe
AxifvEwatKWjeNoa1WqFNAAklNy/FGGGU6LnlNbLDhRHMPjJVFVzgs7FAoGBANcJ
Nuy7b5uVOPumfBLHI1Xns3eHo0POWYl2c1vs3anZppNB3EV1MQykrpU3hokubC6H
bAvef2rMSE9wzhNxBQYKREGlhWne2xNvfLWV2bd3G3tvDc96HXLhXaybnqMgkdRD
2I+t2jnIiOOVRXGF2bl52uIHH1bBrlqM5W+WJl+bAoGBALaYUPB5IW4D9F3wvhij
as5OCjCzBH5lPRfI1EWU4qG/400RonmC61eZlqRALruKfz1alCZ0tSkJX55CqryC
NplouyGcIlz1+muW2GjT7gEOYasJ6Uorep9+mef2RSUvbEX+hx3I57O5U5s70Yt2
EUYy6Evtw5VZzMWO1WodiE5VAoGAAMIIjocOmqbI/6ITl+FZz4i0ijxRKAEHMcPY
Hj/UfC/HNYeq5hfGp3vBHceHUt52BSf3CoerPU4hBx6nq0vfr6jDmtOhh8EAVq4y
61Tu4oWp9CJtEwkkJ26B7QTTZ1HLEct3bPI47bE2Qk8ZYpANN2kli1xpEN435hvP
BzipQ/cCgYAoqQ3mzetmq269RcxfsTBkomiWCzAToTQkSzsHA/0xgB6USl0hS7QU
nXCBBByQvWA9o+aXjAw78hV2A0l1HFfTOM9hfuhDn7LPhYbxW8XGiwcxFxrOZcbH
xjPlYOmP9CJrLdnqWaF0nkUHToq8eI/bpu47AwLMofA5ClJR66tRBw==
-----END RSA PRIVATE KEY-----`

const testingPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwUlXX2eDngAIDlH8XX1N
4kCff98Ev2ulCXMpu1wvR7LrmZ5GQpPtt5VCr7mcy0C8tUe8UZli/jYbicdAYGAG
b8t4d3TqDpewx9igSvWTLMwvZi7+f0fudpdUA3cFZOpKiWdOQSlbIKuf3IAgP/g1
ETb46mOjSSsd7b8gL17m3xUZMnzavGujM4H6t3tTKLnn2CFLmlxXyk/El/FQohQp
s6ihr+MBQxwwnVY9BT0vPxXf2pw5hoHgzqsayT5IDGkqCB3h0Ng6QWumNfzqiJZD
eodNhuW80MZjCThgqC1F2yZukvcMBWHLYnmnYIFZaOPEFuzW9OLJZMdc+XCXmixM
RwIDAQAB
-----END PUBLIC KEY-----`

const testBody = `{"user":"test","country":"BR","currency":"BTC","operator_id":1,"token":"d5dcb4f6-db06-4ab1-9c63-d1f7dc12694e","game_id":10,"lang":"en","lobby_url":"https://casinosample.com","deposit_url":"https://casinosample.com/deposit"}`
const expectedSig = "eLAVF38xYGp4KyHnybr3vEkKY9l7+G3VJusrY5k+FbseYBXZsRcH6EDP9PTUCkAQNftnq61HAo3fyjPKPxJpRQwBJjZpvSHgf5K/VLquf3GU92kXxQwVC1UQzhboKSfk9Ub9tRRt0sQlfRUdYtLSWYDRrWGuDElTOAgE5uhR6Mlkc5UXSKye1JaQHPxrJyVryHCTgFJd+HCy2QYVMQeEl7yF6RYRqmQPGZuawTbvTvz8nRVu5/z5zFFmEHZc2MPQMQAuweP28FjaGnljMWUE89KH5PxiY5CAYZJmez2WXoL9/Voc4c3PJjntAlIOEzLLQ26NNGRKMLwjUsq8ScEEmA=="

func Test_Sign(t *testing.T) {
	sut, err := NewSigner([]byte(testingPrivateKey))
	assert.NoError(t, err)
	signature, err := sut.Sign([]byte(testBody))
	assert.NoError(t, err)

	assert.Equal(t, string(signature), expectedSig)
}

func Test_Verify(t *testing.T) {
	sut, err := NewVerifier([]byte(testingPublicKey))
	assert.NoError(t, err)
	err = sut.Verify(expectedSig, []byte(testBody))
	// no error means the public key could verify the signature was correct
	assert.NoError(t, err)
}

func Test_Verify_Fails_With_Empty_Signature(t *testing.T) {
	sut, err := NewVerifier([]byte(testingPublicKey))
	assert.NoError(t, err)
	err = sut.Verify("", []byte(testBody))
	assert.Error(t, err)
}

func Test_Verify_Fails_With_Wrong_Signature(t *testing.T) {
	sut, err := NewVerifier([]byte(testingPublicKey))
	assert.NoError(t, err)
	err = sut.Verify("fakeSignature", []byte(testBody))
	assert.Error(t, err)
}

func BenchmarkSign(b *testing.B) {
	s, err := NewSigner([]byte(testingPrivateKey))
	assert.NoError(b, err)
	bs := []byte(testBody)
	for i := 0; i < b.N; i++ {
		_, _ = s.Sign(bs)
	}
}

func BenchmarkVerify(b *testing.B) {
	s, err := NewVerifier([]byte(testingPublicKey))
	assert.NoError(b, err)
	payload := []byte(testBody)
	for i := 0; i < b.N; i++ {
		_ = s.Verify(expectedSig, payload)
	}
}
