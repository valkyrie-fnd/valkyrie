package docs

import (
	_ "embed"
	"regexp"

	"github.com/swaggo/swag"
)

// embeds operator_api.yml into docTemplateOperator
//
//go:embed operator_api.yml
var docTemplateOperator string

var SwaggerInfoOperator = &swag.Spec{
	InfoInstanceName: "operator",
	SwaggerTemplate:  docTemplateOperator,
}

func init() {
	// Remove servers property. Should use same as it is being hosted on
	r := regexp.MustCompile(`servers:(.|\s\s\s)*`)
	SwaggerInfoOperator.SwaggerTemplate = r.ReplaceAllString(SwaggerInfoOperator.SwaggerTemplate, "")
	swag.Register(SwaggerInfoOperator.InstanceName(), SwaggerInfoOperator)
}
