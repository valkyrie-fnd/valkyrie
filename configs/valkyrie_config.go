package configs

import (
	"os"
	"reflect"
	"time"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

// LogConfig configuration setup for logging
type LogConfig struct {
	Level  string          `yaml:"level" default:"info"`
	Output OutputLogConfig `yaml:"output"`
	Async  AsyncLogConfig  `yaml:"async"`
}

// AsyncLogConfig Configuration for asynchronous logging
type AsyncLogConfig struct {
	Enabled      bool          `yaml:"enabled" default:"true"`
	BufferSize   int           `yaml:"buffer_size" default:"100000"`
	PollInterval time.Duration `yaml:"poll_interval" default:"10ms"`
}

type OutputLogConfig struct {
	// Type configures where to output logs.
	// Supported types: "stdout", "stderr", "file"
	Type string `yaml:"type" default:"stdout"`

	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.
	Filename string `yaml:"filename,omitempty"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `yaml:"max_size,omitempty"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename. The default is not to remove old log files
	// based on age.
	MaxAge int `yaml:"max_age,omitempty"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `yaml:"max_backups,omitempty"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `yaml:"compress,omitempty"`
}

// TraceConfig Configuration setup for tracing
type TraceConfig struct {
	TraceType       string  `yaml:"type,omitempty"`
	URL             string  `yaml:"url,omitempty"`
	ServiceName     string  `yaml:"service_name,omitempty"`
	GoogleProjectID string  `yaml:"google_project_id,omitempty"`
	SampleRatio     float64 `yaml:"sample_ratio" default:"0.01"`
}

// ProviderConf Configuration structure for provider
type ProviderConf struct {
	// Name of the provider
	Name string `yaml:"name"`
	// Auth Authorization configuration for a specific provider.
	Auth map[string]any `yaml:"auth"`
	// ProviderSpecific Any other config specific to each provider
	ProviderSpecific map[string]any `yaml:"provider_specific,omitempty"`
	// URL url to use for example gamelaunch
	URL string `yaml:"url"`
	// BasePath used to distinguish endpoints exposed by Valkyrie
	BasePath string `yaml:"base_path"`
}

// PamConf Configured information for the used Player Account Manager/wallet
type PamConf = map[string]any

// ValkyrieConfig Parsed valkyrie configuration
type ValkyrieConfig struct {
	HTTPServer       HTTPServerConfig `yaml:"http_server"`
	Pam              PamConf          `yaml:"pam"`
	Tracing          TraceConfig      `yaml:"tracing,omitempty"`
	Providers        []ProviderConf   `yaml:"providers,flow"`
	OperatorBasePath string           `yaml:"operator_base_path"`
	ProviderBasePath string           `yaml:"provider_base_path"`
	Version          string           `yaml:"-"`
	Logging          LogConfig        `yaml:"logging,omitempty"`
	HTTPClient       HTTPClientConfig `yaml:"http_client"`
}

// HTTPServerConfig Configuration used for valkyrie servers
type HTTPServerConfig struct {
	// ProviderAddress configures host and port where Valkyrie will attempt to listen for incoming traffic
	// for provider endpoints.
	//
	// For example, ":8083" binds to all interfaces on port 8083, while "localhost:8083" only
	// binds to local interfaces (no external traffic).
	ProviderAddress string `yaml:"provider_address" default:":8083"`

	// OperatorAddress configures host and port where Valkyrie will attempt to listen for incoming traffic
	// for operator endpoints.
	OperatorAddress string        `yaml:"operator_address" default:":8084"`
	ReadTimeout     time.Duration `yaml:"read_timeout" default:"3s"`  // The amount of time allowed to read the full request including body
	WriteTimeout    time.Duration `yaml:"write_timeout" default:"3s"` // The maximum duration before timing out writes of the response
	IdleTimeout     time.Duration `yaml:"idle_timeout" default:"30s"` // The maximum amount of time to wait for the next request when keep-alive is enabled
}

// HTTPClientConfig Configuration for outgoing requests
type HTTPClientConfig struct {
	ReadTimeout    time.Duration `yaml:"read_timeout" default:"10s"`    // Maximum duration for full response reading (including body)
	WriteTimeout   time.Duration `yaml:"write_timeout" default:"3s"`    // Maximum duration for full request writing (including body)
	RequestTimeout time.Duration `yaml:"request_timeout" default:"10s"` // Maximum duration to wait for the request response (on timeout request will continue in background, try setting read/write timeout to interrupt actual request)
	IdleTimeout    time.Duration `yaml:"idle_timeout" default:"30s"`    // Idle keep-alive connections are closed after this duration.
}

// Read reads yaml file at provided location and parse it into a `ValkyrieConfig`
func Read(file *string) (*ValkyrieConfig, error) {
	data, err := os.ReadFile(*file)
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func parse(data []byte) (*ValkyrieConfig, error) {
	conf := ValkyrieConfig{}

	// Set "default"-tagged values
	if err := defaults.Set(&conf); err != nil {
		return &conf, err
	}

	err := yaml.Unmarshal(data, &conf)

	// Replace environment variables in strings
	expandEnvVariables(&conf, os.ExpandEnv)

	return &conf, err
}

// replaceEnvVariables walks the struct replacing all string fields
// using `os.ExpandEnv()`. Please keep call away from critical path
// for performance reasons.
func expandEnvVariables(strukt any, fn func(string) string) {
	r := reflect.ValueOf(strukt)
	if r.Kind() != reflect.Ptr {
		return
	}
	v := r.Elem()
	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.String:
		if v.CanSet() && !v.IsZero() {
			v.SetString(fn(v.String()))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.CanAddr() || !field.IsValid() || !field.Addr().CanInterface() {
				continue
			}
			expandEnvVariables(field.Addr().Interface(), fn)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			arrVal := v.Index(i)
			if !arrVal.CanAddr() || !arrVal.IsValid() || !arrVal.Addr().CanInterface() {
				continue
			}
			expandEnvVariables(arrVal.Addr().Interface(), fn)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			vv := v.MapIndex(k)
			switch vv.Kind() {
			case reflect.String:
				v.SetMapIndex(k, reflect.ValueOf(fn(vv.String())))
			case reflect.Interface:
				if s, ok := vv.Interface().(string); ok {
					v.SetMapIndex(k, reflect.ValueOf(fn(s)))
				} else {
					expandEnvVariables(vv, fn)
				}
			default:
				expandEnvVariables(vv, fn)
			}
		}
	}
}
