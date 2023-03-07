package caleta_test

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valkyrie-fnd/valkyrie/provider/caleta"
)

var sampleTime = "2023-01-17 08:13:17.985795+00:00"
var sampleTimeWithQuotes = `"2023-01-17 08:13:17.985795+00:00"`
var sampleTimeInstance = time.UnixMicro(1673943197985795)

func Test_UnmarshalTextMsgTimestamp(t *testing.T) {
	var x caleta.MsgTimestamp
	err := x.UnmarshalText([]byte(sampleTime))
	assert.NoError(t, err)
	assert.True(t, sampleTimeInstance.Equal(time.Time(x)))

	err = x.UnmarshalText([]byte("complete gibberish"))
	assert.Error(t, err)
}

func Test_UnmarshalJSONMsgTimestamp(t *testing.T) {
	var params caleta.MsgTimestamp
	err := json.Unmarshal([]byte(sampleTimeWithQuotes), &params)
	assert.NoError(t, err)
	assert.True(t, sampleTimeInstance.Equal(time.Time(params)))
}

func Test_UnmarshalWalletbetParams(t *testing.T) {
	var params caleta.WalletbetParams
	err := json.Unmarshal([]byte(`{
		"X-Auth-Signature":"_",
		"X-Msg-Timestamp":"2023-01-17 08:13:17.985795+00:00"
		}`), &params)

	require.NoError(t, err)
	assert.True(t, sampleTimeInstance.Equal(time.Time(*params.XMsgTimestamp)))
}
