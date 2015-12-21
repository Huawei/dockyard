package router

import (
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/handler"
)

func SetRouters(m *macaron.Macaron) {
	//Docker Registry & Hub V1 API
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
			m.Get("/:imageId/ancestry", handler.GetImageAncestryV1Handler)
			m.Get("/:imageId/json", handler.GetImageJSONV1Handler)
			m.Get("/:imageId/layer", handler.GetImageLayerV1Handler)
			m.Put("/:imageId/json", handler.PutImageJSONV1Handler)
			m.Put("/:imageId/layer", handler.PutImageLayerv1Handler)
			m.Put("/:imageId/checksum", handler.PutImageChecksumV1Handler)
		})
	})

	//Docker Registry & Hub V2 API
	m.Group("/v2", func() {
		m.Get("/", handler.GetPingV2Handler)
		m.Head("/:namespace/:repository/blobs/:digest", handler.HeadBlobsV2Handler)
		m.Post("/:namespace/:repository/blobs/uploads", handler.PostBlobsV2Handler)
		m.Patch("/:namespace/:repository/blobs/uploads/:uuid", handler.PatchBlobsV2Handler)
		m.Put("/:namespace/:repository/blobs/uploads/:uuid", handler.PutBlobsV2Handler)
		m.Get("/:namespace/:repository/blobs/:digest", handler.GetBlobsV2Handler)
		m.Put("/:namespace/:repository/manifests/:tag", handler.PutManifestsV2Handler)
		m.Get("/:namespace/:repository/tags/list", handler.GetTagsListV2Handler)
		m.Get("/:namespace/:repository/manifests/:tag", handler.GetManifestsV2Handler)
	})

	//Rkt Registry & Hub API
	//acis discovery responds endpoints
	m.Get("/:namespace/:aciname/?ac-discovery=1", handler.DiscoveryACIHandler)

	//acis fetch
	m.Get("/:namespace/pubkeys.gpg", handler.GetPubkeysHandler)
	m.Get("/ac-fetch/:namespace/:acifile", handler.GetACIHandler)

	//acis push
	m.Group("/ac-push", func() {
		//m.Put("/:namespace/pubkeys.gpg", handler.PutPubkeysHandler)
		m.Post("/uploaded/:servername/:namespace/:acifile", handler.GetUploadEndPointHandler)
		m.Put("/:namespace/manifest/:acifile", handler.PutManifestHandler)
		m.Put("/:namespace/signature/:acifile", handler.PutSignHandler)
		m.Put("/:namespace/aci/:acifile", handler.PutAciHandler)
		m.Post("/:namespace/complete/:acifile", handler.CompleteHandler)
	})
}
