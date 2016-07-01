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

func SetRouters(m *macaron.Macaron) {
	m.Get("/", handler.IndexV1Handler)

	//Docker Registry V1
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

	//Docker Registry V2
	m.Group("/v2", func() {
		m.Get("/", handler.GetPingV2Handler)
		m.Get("/_catalog", handler.GetCatalogV2Handler)

		//user mode: /namespace/repository:tag
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

		//library mode: /repository:tag
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

	//Appc Discovery
	m.Group("/appc", func() {
		m.Group("/v1", func() {
			//Discovery
			m.Get("/:namespace/:repository/?ac-discovery=1", handler.AppcDiscoveryV1Handler)

			//Pull
			m.Get("/:namespace/:repository/pubkeys", handler.AppcGetPubkeysV1Handler)
			m.Get("/:namespace/:repository/:os/:arch/:aci", handler.AppcGetACIV1Handler)
		})
	})

	//Software Discovery

	//VM Image Discovery

	//Management APIs
}
