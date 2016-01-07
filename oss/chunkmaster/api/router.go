package api

import (
	"net/http"
)

var RouteMap map[string]map[string]http.HandlerFunc

func init() {
	RouteMap = map[string]map[string]http.HandlerFunc{
		"POST": {
			"/cm/v1/chunkserver/batchinitserver": batchInitChunkserverHandler,
			"/cm/v1/chunkserver/initserver":      initChunkserverHandler,
			"/cm/v1/chunkserver/reloadinfo":      loadChunkserverInfoHandler,
			"/cm/v1/chunkserver/reportinfo":      reportChunkserverInfoHandler,
		},
		"GET": {
			"/cm/v1/chunkmaster/route": chunkmasterRouteHandler,
			"/cm/v1/chunkmaster/fid":   chunkmasterFidHandler,

			"/cm/v1/chunkserver/{groupId}/groupinfo": chunkserverGroupInfoHandler,
			"/cm/v1/chunkserver/checkerror":          chunkserverCheckError,
		},
	}
}
