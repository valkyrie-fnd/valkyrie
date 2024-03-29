module github.com/valkyrie-fnd/valkyrie

go 1.21

// You can run a local version of valkyrie-stubs by adding the replace directive like so:
// replace github.com/valkyrie-fnd/valkyrie-stubs => ../valkyrie-stubs

require (
	github.com/creasty/defaults v1.7.0
	github.com/four-fingers/oapi-codegen v0.0.0-20221219135408-9237c9743c67
	github.com/go-playground/validator/v10 v10.15.0
	github.com/goccy/go-json v0.10.2
	github.com/gofiber/contrib/otelfiber v1.0.10
	github.com/gofiber/fiber/v2 v2.48.0
	github.com/gofiber/swagger v0.1.9
	github.com/google/go-querystring v1.1.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-hclog v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.30.0
	github.com/shopspring/decimal v1.3.1
	github.com/stretchr/testify v1.8.4
	github.com/swaggo/swag v1.8.10
	github.com/valkyrie-fnd/valkyrie-stubs v0.0.0-20230814071422-a3d4a4d1cf73
	github.com/valyala/fasthttp v1.48.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.42.0
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v0.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.16.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.39.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
	go.opentelemetry.io/otel/sdk/metric v0.39.0
	go.opentelemetry.io/otel/trace v1.16.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/four-fingers/oapi-codegen-runtime v0.0.0-20230125082134-9d9fdf1239ab // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/hashicorp/go-plugin v1.4.8
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	github.com/swaggo/files v1.0.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	go.opentelemetry.io/contrib v1.17.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.16.0
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	google.golang.org/grpc v1.55.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
)

// avoids depending on all of oapi-codegen's and swag's dependencies
replace (
	github.com/four-fingers/oapi-codegen => github.com/four-fingers/oapi-codegen-runtime v0.1.0
	github.com/swaggo/swag => github.com/four-fingers/swag-runtime v0.1.0
)
