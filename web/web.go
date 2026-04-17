package web

import "embed"

// AdminAssets holds the production build of the Intellectual Admin Dashboard.
//go:embed all:admin/dist
var AdminAssets embed.FS
