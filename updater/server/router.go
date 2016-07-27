/*
Copyright 2016 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"gopkg.in/macaron.v1"

	h "github.com/containerops/dockyard/updater/server/handler"
)

// Dockyard Updater Server Router Definition
func SetRouters(m *macaron.Macaron) {
	// Web API
	m.Get("/", h.IndexMetaV1Handler)

	// App Discovery
	m.Group("/app", func() {
		m.Group("/v1", func() {
			m.Group("/:namespace/:repository", func() {
				// List files
				m.Get("/", h.AppListFileV1Handler)
				// Get meta data of the whole repo
				m.Get("/meta", h.AppGetMetaV1Handler)
				// Get meta signature data of the whole repo
				m.Get("/metasign", h.AppGetMetaSignV1Handler)
				// Get file data of a certain app
				m.Get("/blob/:name", h.AppGetFileV1Handler)
				// Add file to the repo
				m.Post("/:name", h.AppPostFileV1Handler)
			})
		})
	})

}
