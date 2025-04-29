package info

import (
	"galcone/src/app"
	rest "galcone/src/galcone/rest/common"
)

var Router = []*app.RestEndpoint{
	rest.GET("/info", GetInfoHandler),
}
