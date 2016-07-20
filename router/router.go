/*
Copyright 2015 The ContainerOps Authors All rights reserved.

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

package router

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/handler"
)

// Dockyard Router Definition
func SetRouters(m *macaron.Macaron) {
	// Web API
	m.Get("/", handler.IndexV1Handler)

	// Docker Registry V1
	m.Group("/v1", func() {
		m.Get("/_ping", handler.GetPingV1Handler)

		m.Get("/users", handler.GetUsersV1Handler)
		m.Post("/users", handler.PostUsersV1Handler)

		m.Group("/repositories", func() {
			m.Put("/:namespace/:repository/tags/:tag", handler.PutTagV1Handler)
			m.Put("/:namespace/:repository/images", handler.PutRepositoryImagesV1Handler)
			m.Get("/:namespace/:repository/images", handler.GetRepositoryImagesV1Handler)
			m.Get("/:namespace/:repository/tags", handler.GetTagV1Handler)
			m.Put("/:namespace/:repository", handler.PutRepositoryV1Handler)
		})

		m.Group("/images", func() {
			m.Get("/:image/ancestry", handler.GetImageAncestryV1Handler)
			m.Get("/:image/json", handler.GetImageJSONV1Handler)
			m.Get("/:image/layer", handler.GetImageLayerV1Handler)
			m.Put("/:image/json", handler.PutImageJSONV1Handler)
			m.Put("/:image/layer", handler.PutImageLayerV1Handler)
			m.Put("/:image/checksum", handler.PutImageChecksumV1Handler)
		})
	})

	// Docker Registry V2
	m.Group("/v2", func() {
		m.Get("/", handler.GetPingV2Handler)
		m.Get("/_catalog", handler.GetCatalogV2Handler)

		// user mode: /namespace/repository:tag
		m.Head("/:namespace/:repository/blobs/:digest", handler.HeadBlobsV2Handler)
		m.Post("/:namespace/:repository/blobs/uploads", handler.PostBlobsV2Handler)
		m.Patch("/:namespace/:repository/blobs/uploads/:uuid", handler.PatchBlobsV2Handler)
		m.Put("/:namespace/:repository/blobs/uploads/:uuid", handler.PutBlobsV2Handler)
		m.Get("/:namespace/:repository/blobs/:digest", handler.GetBlobsV2Handler)
		m.Put("/:namespace/:repository/manifests/:tag", handler.PutManifestsV2Handler)
		m.Get("/:namespace/:repository/tags/list", handler.GetTagsListV2Handler)
		m.Get("/:namespace/:repository/manifests/:tag", handler.GetManifestsV2Handler)
		m.Delete("/:namespace/:repository/blobs/:digest", handler.DeleteBlobsV2Handler)
		m.Delete("/:namespace/:repository/manifests/:reference", handler.DeleteManifestsV2Handler)

		// library mode: /repository:tag
		m.Head("/:repository/blobs/:digest", handler.HeadBlobsV2LibraryHandler)
		m.Post("/:repository/blobs/uploads", handler.PostBlobsV2LibraryHandler)
		m.Patch("/:repository/blobs/uploads/:uuid", handler.PatchBlobsV2LibraryHandler)
		m.Put("/:repository/blobs/uploads/:uuid", handler.PutBlobsV2LibraryHandler)
		m.Get("/:repository/blobs/:digest", handler.GetBlobsV2LibraryHandler)
		m.Put("/:repository/manifests/:tag", handler.PutManifestsV2LibraryHandler)
		m.Get("/:repository/tags/list", handler.GetTagsListV2LibraryHandler)
		m.Get("/:repository/manifests/:tag", handler.GetManifestsV2LibraryHandler)
		m.Delete("/:repository/blobs/:digest", handler.DeleteBlobsV2LibraryHandler)
		m.Delete("/:repository/manifests/:reference", handler.DeleteManifestsV2LibraryHandler)
	})

	// App Discovery
	m.Group("/app", func() {
		m.Group("/v1", func() {
			// Global Search
			m.Get("/search", handler.AppGlobalSearchV1Handler)

			m.Group("/:namespace/:repository", func() {
				// Discovery
				m.Get("/?app-discovery=1", handler.AppDiscoveryV1Handler)

				// Scoped Search
				m.Get("/search", handler.AppScopedSearchV1Handler)
				m.Get("/list", handler.AppGetListAppV1Handler)

				// Pull
				m.Get("/:os/:arch/:app", handler.AppGetFileV1Handler)

				// Push
				m.Post("/", handler.AppPostV1Handler)
				m.Put("/:os/:arch/:app/:tag", handler.AppPutFileV1Handler)
				m.Put("/:os/:arch/:app/:tag/manifests", handler.AppPutManifestV1Handler)
				m.Patch("/:os/:arch/:app/:tag/:status", handler.AppPatchFileV1Handler)
				m.Delete("/:os/:arch/:app/:tag", handler.AppDeleteFileByTagV1Handler)
				m.Delete("/:os/:arch/:app", handler.AppDeleteFileV1Handler)
			})
		})
	})

	// Appc Discovery
	m.Group("/appc", func() {
		m.Group("/v1", func() {
			m.Group("/:namespace/:repository", func() {
				// Discovery
				m.Get("/?ac-discovery=1", handler.AppcDiscoveryV1Handler)

				// Pull
				m.Get("/pubkeys", handler.AppcGetPubkeysV1Handler)
				m.Get("/:os/:arch/:aci", handler.AppcGetACIV1Handler)
			})

		})
	})

	// VM Image Discovery
	m.Group("/image", func() {
		m.Group("/v1", func() {
			// Global Search
			m.Get("/search", handler.ImageGlobalSearchV1Handler)

			m.Group("/:namespace/:repository", func() {
				// Discovery
				m.Get("/?image-discovery=1", handler.ImageDiscoveryV1Handler)

				// Scoped Search
				m.Get("/search", handler.ImageScopedSearchV1Handler)
				m.Get("/list", handler.ImageGetListV1Handler)

				// Pull
				m.Get("/:os/:arch/:image", handler.ImageGetFileV1Handler)

				// Push
				m.Post("/", handler.ImagePostV1Handler)
				m.Put("/:os/:arch/:image/:tag", handler.ImagePutFileV1Handler)
				m.Put("/:os/:arch/:image/:tag/manifests", handler.ImagePutManifestV1Handler)
				m.Patch("/:os/:arch/:image/:tag/:status", handler.ImagePatchFileV1Handler)
				m.Delete("/:os/:arch/:image/:tag", handler.ImageDeleteFileByTagV1Handler)
				m.Delete("/:os/:arch/:image", handler.ImageDeleteFileV1Handler)
			})
		})
	})

	// Sync APIS
	m.Group("/sync", func() {
		m.Group("/v1", func() {
			// Server Ping
			m.Get("/ping", handler.SyncGetPingV1Handler)

			m.Group("/master", func() {
				// Server Sync Of Master
				m.Post("/registry", handler.SyncMasterPostRegistryV1Handler)
				m.Delete("/registry", handler.SyncMasterDeleteRegistryV1Handler)

				m.Put("/mode", handler.SyncMasterPutModeRegistryV1Handler)
			})

			m.Group("/slave", func() {
				// Server Sync Of Slaver
				m.Post("/registry", handler.SyncSlavePostRegistryV1Handler)
				m.Put("/registry", handler.SyncSlavePutRegistryV1Handler)
				m.Delete("/registry", handler.SyncSlaveDeleteRegistryV1Handler)

				m.Put("/mode", handler.SyncSlavePutModeRegistryV1Handler)

				// Data Sync
				m.Get("/list", handler.SyncSlaveListDataV1Handler)

				// File Sync
				m.Put("/:namespace/:repository/manifests", handler.SyncSlavePutManifestsV1Handler)
				m.Put("/:namespace/:repository/file", handler.SyncSlavePutFileV1Handler)
				m.Put("/:namespace/:repository/:status", handler.SyncSlavePutStatusV1Handler)
			})
		})
	})

	// Admin APIs
	m.Group("/admin", func() {
		m.Group("/v1", func() {
			// Server Status
			m.Get("/stats/:type", handler.AdminGetStatusV1Handler)

			// Server Config
			m.Get("/config", handler.AdminGetConfigV1Handler)
			m.Put("/config", handler.AdminSetConfigV1Handler)

			// Maintenance
			m.Post("/maintenance", handler.AdminPostMaintenance)
		})
	})
}
