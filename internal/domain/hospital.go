// Package domain holds the entities shared by every layer and the errors they exchange.
package domain

// Hospital is a hospital the middleware serves. Code is the natural key used everywhere.
type Hospital struct {
	Code string
	Name string
}
