package Services

import "github.com/AdamElHassanLeb/VOD-Downloader/API/pkg/Structs"

type Service struct {
	VODIngest interface {
		IngestHLS(request Structs.VODIngestRequest) error
	}
}

func NewService() Service {
	return Service{
		VODIngest: &VODIngestService{},
	}
}
