// Package swagger adds API documentation endpoints
package swagger

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/rs/zerolog/log"
	"github.com/swaggo/swag"

	"github.com/valkyrie-fnd/valkyrie/provider/docs/generated"

	"github.com/valkyrie-fnd/valkyrie/configs"

	_ "github.com/valkyrie-fnd/valkyrie/provider/docs"
)

// @title           Valkyrie Provider API
// @host            localhost:8083
// @basePath        /
// @schemes         http
// @version         -
// @description     The exposed endpoints by the enabled game provider modules.
func ConfigureSwagger(config *configs.ValkyrieConfig, provider, operator fiber.Router) error {
	log.Info().Msg("Registering Swagger routes under GET /swagger/")

	// Get tags to filter generated documentation on
	tags := getTags(config)

	generated.SwaggerInfoprovider.Version = config.Version
	generated.SwaggerInfoprovider.Host = config.HTTPServer.ProviderAddress

	// Filter documentation (containing all providers) to only include the tags we want (configured providers)
	err := postProcessing(generated.SwaggerInfoprovider, tags)
	if err != nil {
		return err
	}

	provider.Get("/swagger/*", swagger.New(swagger.Config{
		InstanceName: "provider",
	}))

	operator.Get("/swagger/*", swagger.New(swagger.Config{
		InstanceName: "operator",
	}))

	return nil
}

func getTags(config *configs.ValkyrieConfig) map[string]struct{} {
	providers := map[string]struct{}{}
	for _, v := range config.Providers {
		providers[v.Name] = struct{}{}
	}
	return providers
}

// postProcessing filters out documentation for the tags specified in keepTags
func postProcessing(spec *swag.Spec, keepTags map[string]struct{}) error {
	// Generated spec had invalid json in it, lets fix before unmarshalling
	spec.SwaggerTemplate = strings.ReplaceAll(spec.SwaggerTemplate, "\"schemes\": {{ marshal .Schemes }},", "\"schemes\": \"{{ marshal .Schemes }}\",")

	// Read the template into a generic map
	var template map[string]any
	err := json.Unmarshal([]byte(spec.SwaggerTemplate), &template)
	if err != nil {
		return err
	}

	filterOperations(template, keepTags)
	filterDefinitions(template, keepTags)

	// Write the post processed template back to JSON
	bytes, err := json.Marshal(&template)
	if err != nil {
		return err
	}
	spec.SwaggerTemplate = string(bytes)

	// oh, generated spec needs the invalid json back, otherwise templating breaks
	spec.SwaggerTemplate = strings.ReplaceAll(spec.SwaggerTemplate, "\"schemes\":\"{{ marshal .Schemes }}\",", "\"schemes\": {{ marshal .Schemes }},")

	return nil
}

// Filter definitions with the tags we want
func filterDefinitions(template map[string]any, keepTags map[string]struct{}) {
	// Same goes for definitions
	definitions := template["definitions"].(map[string]any)
	for k := range definitions {
		remove := true
		for tag := range keepTags {
			tag = strings.ToLower(tag)
			tag = strings.ReplaceAll(tag, " ", "")
			if strings.HasPrefix(k, tag) {
				remove = false
			}
		}
		if remove {
			delete(definitions, k)
		}
	}
}

// Filter operations with the tags we want
func filterOperations(template map[string]any, keepTags map[string]struct{}) {
	paths := template["paths"].(map[string]any)
	for path, m := range paths {
		methods := m.(map[string]any)
		for method, o := range methods {
			operation := o.(map[string]any)
			tags := operation["tags"].([]any)

			var newTags []string
			for _, t := range tags {
				tag := t.(string)
				if _, found := keepTags[tag]; found {
					newTags = append(newTags, tag)
				}
			}
			operation["tags"] = newTags

			if len(newTags) == 0 {
				delete(methods, method)
			}
		}
		if len(methods) == 0 {
			delete(paths, path)
		}
	}
}
