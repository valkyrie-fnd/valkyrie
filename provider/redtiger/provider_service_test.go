package redtiger

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valkyrie-fnd/valkyrie/configs"
	"github.com/valkyrie-fnd/valkyrie/provider"
)

type GLTestData struct {
	name    string
	e       error
	want    string
	conf    *configs.ProviderConf
	req     *provider.GameLaunchRequest
	headers *provider.GameLaunchHeaders
}

var gameLaunchTests = []GLTestData{
	{
		name:    "Error when request is missing Session key",
		want:    "",
		e:       errors.New("Missing SessionKey"),
		conf:    &configs.ProviderConf{},
		req:     &provider.GameLaunchRequest{},
		headers: &provider.GameLaunchHeaders{},
	},
	{
		name: "Error when game launch is missing playMode",
		want: "",
		e:    errors.New("Key: 'rtGameLaunchConfig.PlayMode' Error:Field validation for 'PlayMode' failed on the 'required' tag"),
		conf: &configs.ProviderConf{},
		req: &provider.GameLaunchRequest{
			Currency: "USD",
			PlayerID: "1",
		},
		headers: &provider.GameLaunchHeaders{SessionKey: "123"},
	},
	{
		name: "Returns gameurl with queryparameters",
		want: fmt.Sprintf(
			"http://rt-baseUrl.com/GameId123?%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s",
			"currency", "USD",
			"token", "123",
			"userId", "1",
			"fullScreen", "true",
			"hasAutoplayLimitLoss", "false",
			"hasAutoplaySingleWinLimit", "false",
			"hasAutoplayStopOnBonus", "false",
			"hasAutoplayStopOnJackpot", "false",
			"hasAutoplayTotalSpins", "false",
			"hasFreeBets", "false",
			"hasHistory", "false",
			"hasRealPlayButton", "false",
			"hasRoundId", "false",
			"playMode", "real",
		),
		e: nil,
		conf: &configs.ProviderConf{
			URL: "http://rt-baseUrl.com",
		},
		req: &provider.GameLaunchRequest{
			Currency:       "USD",
			PlayerID:       "1",
			ProviderGameID: "GameId123",
			LaunchConfig: map[string]interface{}{
				"PlayMode":   "real",
				"FullScreen": true,
			},
		},
		headers: &provider.GameLaunchHeaders{SessionKey: "123"},
	},
}

func TestGameLaunch(t *testing.T) {
	for _, test := range gameLaunchTests {
		t.Run(test.name, func(t *testing.T) {
			sut := RedTigerService{test.conf}
			result, err := sut.GameLaunch(nil, test.req, test.headers)
			if test.e != nil {
				assert.EqualError(t, err, test.e.Error())
				assert.Equal(t, "", result)
			}
			if result != "" {
				assert.Equal(t, test.want, result)
			}
		})
	}
}

func TestGameRoundRender(t *testing.T) {
	sut := RedTigerService{}
	_, err := sut.GetGameRoundRender(nil, provider.GameRoundRenderRequest{})
	assert.EqualError(t, err, "Not available")
}
