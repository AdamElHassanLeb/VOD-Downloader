package Controllers

import "github.com/AdamElHassanLeb/VOD-Downloader/API/cmd/internal/Services"

var Service Services.Service

func init() {
	Service = Services.NewService()
}
