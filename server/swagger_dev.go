//go:build dev

package server

import "github.com/valkyrie-fnd/valkyrie/swagger"

// configureSwagger includes and configures swagger when a dev build
func configureSwagger(v *Valkyrie) error {
	return swagger.ConfigureSwagger(v.config, v.provider, v.operator)
}
