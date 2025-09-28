package Services

import (
	"context"
	"github.com/AdamElHassanLeb/VOD-Downloader/API/pkg/Env"
	"github.com/AdamElHassanLeb/VOD-Downloader/API/pkg/Structs"
	"github.com/AdamElHassanLeb/VOD-Downloader/API/pkg/VODIngestors"
	"sync"
)

type VODIngestService struct{}

var ingestorMap sync.Map = sync.Map{}

func (VIS *VODIngestService) IngestHLS(request Structs.VODIngestRequest) error {

	hlsIngestor, err := VODIngestors.NewHLSVODIngestor(
		context.Background(),
		request.Name,
		request.Episode,
		request.Url,
		Env.GetString("VOD_DIR", "./VOD_DIR"),
		&ingestorMap)

	if err != nil {
		return err
	}
	ingestorMap.Store(request.Name+"EP"+request.Episode, hlsIngestor)
	go hlsIngestor.Start()

	return nil
}
