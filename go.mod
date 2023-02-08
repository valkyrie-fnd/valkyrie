module github.com/valkyrie-fnd/valkyrie

go 1.19

// You can run a local version of valkyrie-stubs by adding the replace directive like so:
// replace github.com/valkyrie-fnd/valkyrie-stubs => ../valkyrie-stubs

require (
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.11.0
	github.com/creasty/defaults v1.6.0
	github.com/four-fingers/oapi-codegen v0.0.0-20221219135408-9237c9743c67
	github.com/go-playground/validator/v10 v10.11.2
	github.com/goccy/go-json v0.10.0
	github.com/gofiber/contrib/otelfiber v0.0.0-20230208131514-990d42830886
	github.com/gofiber/fiber/v2 v2.42.0
	github.com/gofiber/swagger v0.1.9
	github.com/google/go-querystring v1.1.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-hclog v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.29.0
	github.com/shopspring/decimal v1.3.1
	github.com/stretchr/testify v1.8.1
	github.com/swaggo/swag v1.8.10
	github.com/valkyrie-fnd/valkyrie-stubs v0.0.0-20230207114947-2b0da4b82e50
	github.com/valyala/fasthttp v1.44.0
	go.opentelemetry.io/otel v1.13.0
	go.opentelemetry.io/otel/exporters/jaeger v1.13.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.13.0
	go.opentelemetry.io/otel/sdk v1.13.0
	go.opentelemetry.io/otel/trace v1.13.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute v1.15.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/trace v1.8.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.35.0 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/four-fingers/oapi-codegen-runtime v0.0.0-20230125082134-9d9fdf1239ab // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.1 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	github.com/hashicorp/go-plugin v1.4.8
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/klauspost/compress v1.15.15 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/savsgio/dictpool v0.0.0-20221023140959-7bf2e61cea94 // indirect
	github.com/savsgio/gotils v0.0.0-20230203094617-bcbc01813b4f // indirect
	github.com/swaggo/files v1.0.0 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib v1.13.0 // indirect
	go.opentelemetry.io/otel/metric v0.36.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/oauth2 v0.4.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/api v0.108.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230125152338-dcaf20b6aeaa // indirect
	google.golang.org/grpc v1.52.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

require github.com/kr/text v0.2.0 // indirect

// avoids depending on all of oapi-codegen's and swag's dependencies
replace (
	github.com/four-fingers/oapi-codegen => github.com/four-fingers/oapi-codegen-runtime v0.1.0
	github.com/swaggo/swag => github.com/four-fingers/swag-runtime v0.1.0
)
