//go:build !dev

package server

// configureSwagger does nothing in prod
func configureSwagger(_ *Valkyrie) error {
	return nil
}
