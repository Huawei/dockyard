package models

import (
    "time"
)


type Aci struct {
	Name    string
	Details []Acidetails
}

type Acidetails struct {
	Version string
	OS      string
	Arch    string
	Signed  bool
	LastMod string
}

type InitiateDetails struct {
	ACIPushVersion string `json:"aci_push_version"`
	Multipart      bool   `json:"multipart"`
	ManifestURL    string `json:"upload_manifest_url"`
	SignatureURL   string `json:"upload_signature_url"`
	ACIURL         string `json:"upload_aci_url"`
	CompletedURL   string `json:"completed_url"`
}

type CompleteMsg struct {
	Success      bool   `json:"success"`
	Reason       string `json:"reason,omitempty"`
	ServerReason string `json:"server_reason,omitempty"`
}

type Upload struct {
	Started time.Time
	Image   string
	GotSig  bool
	GotACI  bool
	GotMan  bool
}

