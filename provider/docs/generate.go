package docs

// Run using "go generate ./..."
//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -g ./ops/swagger.go -d ../../ -o ./generated/ --instanceName provider --ot go,yaml
