// Package web is the public Catalog server entry point. HTTP implementation lives
// in internal/server; pure playground selection and presentation data have no
// dependency on mutable server state. Templates and assets have separate owners.
package web

//go:generate go run ../../cmd/generate-templates
