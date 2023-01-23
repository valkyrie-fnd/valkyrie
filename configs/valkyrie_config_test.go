package configs

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	signingKey = `-----BEGIN RSA PRIVATE KEY-----
caleta-signing-key
-----END RSA PRIVATE KEY-----
`
	verificationKey = `-----BEGIN PUBLIC KEY-----
caleta-verification-key
-----END PUBLIC KEY-----
`
)

type testWrapper struct {
	name        string
	envFilePath string
	yamlData    string
	want        *ValkyrieConfig
}

var defaultHTTPServerConfig = HTTPServerConfig{
	ReadTimeout:     3 * time.Second,
	WriteTimeout:    3 * time.Second,
	IdleTimeout:     30 * time.Second,
	ProviderAddress: ":8083",
	OperatorAddress: ":8084",
}

var defaultHTTPClientConfig = HTTPClientConfig{
	ReadTimeout:    10 * time.Second,
	WriteTimeout:   3 * time.Second,
	IdleTimeout:    30 * time.Second,
	RequestTimeout: 10 * time.Second,
}

var defaultLogConfig = LogConfig{
	Level: "info",
	Async: AsyncLogConfig{
		Enabled:      true,
		BufferSize:   100000,
		PollInterval: 10 * time.Millisecond,
	},
	Output: OutputLogConfig{Type: "stdout"},
}

var defaultTraceConfig = TraceConfig{
	SampleRatio: 0.01,
}

var tests = []testWrapper{
	{
		name:        "yaml parsed successfully",
		envFilePath: "",
		yamlData: `
tracing:
  type: jaeger
  url: 'http://localhost'
  service_name: my-service
`,
		want: &ValkyrieConfig{
			Tracing: TraceConfig{
				TraceType:   "jaeger",
				URL:         "http://localhost",
				ServiceName: "my-service",
				SampleRatio: 0.01,
			},
			Logging:    defaultLogConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
		},
	},
	{
		name:        "Yaml with env variables parsed",
		envFilePath: "testdata/zipkin.env",
		yamlData: `
tracing:
  type: ${TRACING_TYPE}
  url: ${TRACING_URL}
  service_name: ${TRACING_SERVICE_NAME}
`,
		want: &ValkyrieConfig{
			Tracing: TraceConfig{
				TraceType:   "zipkin",
				URL:         "the.zipkin.host:222/path",
				ServiceName: "my-service",
				SampleRatio: 0.01,
			},
			Logging:    defaultLogConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
		},
	},
	{
		name:        "yaml parsed with ",
		envFilePath: "",
		yamlData: `
tracing:
  type: jaeger
  url: ${ENV_THAT_DOESNT_EXIST}
  service_name: my-service
`,
		want: &ValkyrieConfig{
			Tracing: TraceConfig{
				TraceType:   "jaeger",
				URL:         "",
				ServiceName: "my-service",
				SampleRatio: 0.01,
			},
			Logging:    defaultLogConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
		},
	},
	{
		name:        "yaml parsed with sampleRatio",
		envFilePath: "",
		yamlData: `
tracing:
  type: jaeger
  service_name: my-service
  sample_ratio: 1.0
`,
		want: &ValkyrieConfig{
			Tracing: TraceConfig{
				TraceType:   "jaeger",
				URL:         "",
				ServiceName: "my-service",
				SampleRatio: 1.0,
			},
			Logging:    defaultLogConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
		},
	},
	{
		name:        "parsed and variables substituted",
		envFilePath: "testdata/some.env",
		yamlData: `
pam:
  name: generic
  api_key: ${KEY1}
  url: 'http://pam.url'
providers:
  - name: providerA
    url: 'https://some.url'
    auth:
      casino_key: ${KEY2}
    `,
		want: &ValkyrieConfig{
			Pam: PamConf{
				"name":    "generic",
				"api_key": "key-one",
				"url":     "http://pam.url",
			},
			Tracing: TraceConfig{SampleRatio: 0.01},
			Providers: []ProviderConf{
				{
					Name: "providerA",
					URL:  "https://some.url",
					Auth: map[string]any{"casino_key": "key-two"},
				},
			},
			Logging:    defaultLogConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
		},
	},
	{
		name:        "providers auth field is parsed to a map[string]string",
		envFilePath: "",
		yamlData: `
pam:
  name: generic
  api_key: 123xyz
  url: 'http://pam.url'
providers:
  - name: providerA
    url: 'https://some.url'
    auth:
      casino_key: xxx
      api_key: someKey
      casino_token: secretToken
      other_key: fooValue`,
		want: &ValkyrieConfig{
			Pam: PamConf{
				"name":    "generic",
				"api_key": "123xyz",
				"url":     "http://pam.url",
			},
			Tracing: TraceConfig{SampleRatio: 0.01},
			Providers: []ProviderConf{
				{
					Name: "providerA",
					URL:  "https://some.url",
					Auth: map[string]any{"casino_key": "xxx", "api_key": "someKey", "casino_token": "secretToken", "other_key": "fooValue"},
				},
			},
			Logging:    defaultLogConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
		},
	},
	{
		name:        "yaml with defaults overridden parsed correctly",
		envFilePath: "",
		yamlData: `
logging:
  level: debug
  async:
    enabled: false
    buffer_size: 50000
    poll_interval: 5ms
http_server:
  read_timeout: 2s
  write_timeout: 100ms
  idle_timeout: 10s
http_client:
  read_timeout: 2s
  write_timeout: 100ms
  idle_timeout: 10s
  request_timeout: 2s
`,
		want: &ValkyrieConfig{
			Logging: LogConfig{
				Level: "debug",
				Async: AsyncLogConfig{
					Enabled:      false,
					BufferSize:   50000,
					PollInterval: 5 * time.Millisecond,
				},
				Output: OutputLogConfig{Type: "stdout"},
			},
			Tracing: TraceConfig{SampleRatio: 0.01},
			HTTPServer: HTTPServerConfig{
				ReadTimeout:     2 * time.Second,
				WriteTimeout:    100 * time.Millisecond,
				IdleTimeout:     10 * time.Second,
				ProviderAddress: ":8083",
				OperatorAddress: ":8084",
			},
			HTTPClient: HTTPClientConfig{
				ReadTimeout:    2 * time.Second,
				WriteTimeout:   100 * time.Millisecond,
				IdleTimeout:    10 * time.Second,
				RequestTimeout: 2 * time.Second,
			},
		},
	},
	{
		name:        "Yaml with multiline env variables",
		envFilePath: "testdata/test-key.env",
		yamlData: `
providers:
- name: some
  auth:
    key: ${TEST_KEY}
`,
		want: &ValkyrieConfig{
			Providers: []ProviderConf{
				{
					Name: "some",
					Auth: map[string]any{"key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA0whbOMM8Kws4EzFl4pmZ6blW1JSe"},
				},
			},
			Logging:    defaultLogConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
			Tracing:    TraceConfig{SampleRatio: 0.01},
		},
	},
	{
		name:        "yaml with file logging output parsed successfully",
		envFilePath: "",
		yamlData: `
logging:
  output:
    type: file
    filename: test
    max_size: 1
    max_age: 2
    max_backups: 3
    compress: true
`,
		want: &ValkyrieConfig{
			Logging: LogConfig{
				Level: "info",
				Async: AsyncLogConfig{
					Enabled:      true,
					BufferSize:   100000,
					PollInterval: 10 * time.Millisecond,
				},
				Output: OutputLogConfig{
					Type:       "file",
					Filename:   "test",
					MaxSize:    1,
					MaxAge:     2,
					MaxBackups: 3,
					Compress:   true,
				},
			},
			Tracing:    defaultTraceConfig,
			HTTPServer: defaultHTTPServerConfig,
			HTTPClient: defaultHTTPClientConfig,
		},
	},
}

func TestValkConfig(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithEnvAndReset(tt.envFilePath, func() {
				cfg, err := parse([]byte(tt.yamlData))
				require.NoError(t, err)
				assert.Equal(t, tt.want, cfg)
			})
		})
	}
}

func TestValkConfigFile(t *testing.T) {
	var file = "testdata/valkyrie_config.test.yml"
	var expectedConfig = &ValkyrieConfig{
		Logging: LogConfig{
			Level: "debug",
			Async: AsyncLogConfig{
				Enabled:      true,
				BufferSize:   500000,
				PollInterval: 5 * time.Millisecond,
			},
			Output: OutputLogConfig{
				Type: "stdout",
			},
		},
		Tracing: TraceConfig{
			TraceType:       "stdout",
			ServiceName:     "traceServiceName",
			GoogleProjectID: "xyz",
			URL:             "https://tracing-server-url",
			SampleRatio:     0.01,
		},
		Pam: PamConf{
			"name":    "generic",
			"api_key": "pam-api-key",
			"url":     "https://pam-url",
		},
		Providers: []ProviderConf{
			{
				Name:     "Evolution",
				URL:      "https://evo-url",
				BasePath: "/evolution",
				Auth:     map[string]any{"api_key": "evo-api-key", "casino_token": "evo-casino-token", "casino_key": "evo-casino-key"},
			},
			{
				Name:     "Red Tiger",
				URL:      "https://rt-url",
				BasePath: "/redtiger",
				Auth:     map[string]any{"api_key": "rt-api-key", "recon_token": "rt-recon-token"},
			},
			{
				Name:     "Caleta",
				URL:      "https://caleta-url",
				BasePath: "/caleta",
				Auth:     map[string]any{"verification_key": verificationKey, "signing_key": signingKey, "operator_id": "caleta-operator-id"},
			},
		},
		HTTPServer: HTTPServerConfig{
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			IdleTimeout:     30 * time.Second,
			ProviderAddress: "localhost:8083",
			OperatorAddress: "localhost:8084",
		},
		HTTPClient: defaultHTTPClientConfig,
	}
	cfg, err := Read(&file)
	require.NoError(t, err)
	assert.Equal(t, expectedConfig, cfg)
}

// Just runs the func with env vars set and then clears the vars
func runWithEnvAndReset(file string, fn func()) {
	if file == "" {
		fn()
		return
	}
	vars, _ := godotenv.Read(file)
	_ = godotenv.Overload(file)

	fn()

	for k := range vars {
		_ = os.Unsetenv(k)
	}
}

func Test_expandEnvVariables(t *testing.T) {

	tests := []struct {
		name     string
		strukt   any
		expected any
	}{
		{
			"paint a plain string",
			strPtr("something"),
			strPtr("pelle"),
		},
		{
			"paint string map",
			&map[string]string{"apa": "apa", "träd": "träd"},
			&map[string]string{"apa": "pelle", "träd": "pelle"},
		},
		{
			"paint struct",
			&struct {
				name string
				Name string
			}{
				name: "apa",
				Name: "apa",
			},
			&struct {
				name string
				Name string
			}{
				name: "apa",
				Name: "pelle",
			},
		},
		{
			"actual config",
			&ValkyrieConfig{
				Logging: LogConfig{Level: "debug"},
			},
			&ValkyrieConfig{
				Logging: LogConfig{Level: "pelle"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test repalces all strings with the same word
			expandEnvVariables(tt.strukt, func(s string) string { return "pelle" })
			assert.Equal(t, tt.expected, tt.strukt)
		})
	}
}

func TestMinimalValkConfigFile(t *testing.T) {
	var file = "testdata/valkyrie_config.minimal.yml"
	_, err := Read(&file)
	require.NoError(t, err)
}

func strPtr(s string) *string {
	return &s
}
