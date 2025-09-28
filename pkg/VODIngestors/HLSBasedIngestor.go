package VODIngestors

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/AdamElHassanLeb/VOD-Downloader/API/pkg/Errors"
	"github.com/grafov/m3u8"
)

type HLSVODIngestor struct {
	URL            *url.URL
	Name           string
	Episode        string
	BaseDir        string
	orginalBaseDir string
	ctx            context.Context
	cancel         context.CancelFunc
	client         *http.Client
	ingestorMap    *sync.Map
}

func NewHLSVODIngestor(parentCtx context.Context, name, episode, urlRaw, location string, ingestorMap *sync.Map) (*HLSVODIngestor, error) {

	u, err := url.Parse(urlRaw)

	if err != nil {
		log.Println(err.Error())
		return nil, Errors.ErrInvalidURL
	}

	usr, _ := user.Current()
	storageDir := filepath.Join(usr.HomeDir, location, name, episode)

	/*
		if _, err := os.Stat(storageDir); !os.IsNotExist(err) {
			return nil, Errors.ErrMediaExists
		}*/

	if err := os.MkdirAll(storageDir, 0777); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(parentCtx)

	return &HLSVODIngestor{
		URL:            u,
		Name:           name,
		Episode:        episode,
		BaseDir:        storageDir,
		orginalBaseDir: location,
		ctx:            ctx,
		cancel:         cancel,
		client: &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: 60 * time.Second,
			},
		},
		ingestorMap: ingestorMap,
	}, nil
}

func (hvi *HLSVODIngestor) Start() {
	log.Printf("Starting HLSVOD Ingestor: %s Episode %s", hvi.Name, hvi.Episode)
	hvi.runLoop()
}

func (hvi *HLSVODIngestor) runLoop() error {
	for {
		select {
		case <-hvi.ctx.Done():
			return nil
		default:
		}

		r, err := hvi.client.Get(hvi.URL.String())

		if err != nil {
			log.Printf("Error fetching playlist %s: %v", hvi.URL, err)
			time.Sleep(2 * time.Second)
			continue
		}
		playlist, plType, err := m3u8.DecodeFrom(r.Body, true)

		r.Body.Close()

		//TODO iterate a val then return error after multiple attempts
		if err != nil {
			log.Printf("Error parsing playlist %s: %v", hvi.Name, err)
			time.Sleep(2 * time.Second)
			continue
		}

		switch plType {

		case m3u8.MASTER:
			master := playlist.(*m3u8.MasterPlaylist)
			log.Printf("Master playlist detected (%d variants)", len(master.Variants))

			for _, variant := range master.Variants {
				rel, err := url.Parse(variant.URI)
				if err != nil {
					continue
				}
				absURL := hvi.URL.ResolveReference(rel).String()

				mediaIng, err := NewHLSVODIngestor(hvi.ctx, hvi.Name, hvi.Episode, absURL, hvi.orginalBaseDir, hvi.ingestorMap)
				if err == nil {
					go mediaIng.Start()
				}
			}
			err := hvi.saveMasterPlaylist(master)
			if err != nil {
				return err
			}
			return nil

		case m3u8.MEDIA:

			media := playlist.(*m3u8.MediaPlaylist)
			go hvi.saveMediaPlaylist(media)

			return nil
		}
	}
}

func (hvi *HLSVODIngestor) saveMasterPlaylist(pl *m3u8.MasterPlaylist) error {
	buf := &bytes.Buffer{}
	pl.Encode().WriteTo(buf)

	diskPath := filepath.Join(hvi.BaseDir, "index.m3u8")
	os.MkdirAll(filepath.Dir(diskPath), 0777)

	if err := os.WriteFile(diskPath, buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving master playlist: %v", err)
		return err
	}
	return nil
}

func (hvi *HLSVODIngestor) saveMediaPlaylist(pl *m3u8.MediaPlaylist) error {

	buf := &bytes.Buffer{}
	_, err2 := pl.Encode().WriteTo(buf)
	if err2 != nil {
		log.Printf("Error saving master playlist: %v", err2)
		return err2
	}

	fname := filepath.Base(hvi.URL.String())

	diskPath := filepath.Join(hvi.BaseDir, fname)

	err := os.MkdirAll(filepath.Dir(diskPath), 0777)
	if err != nil {
		log.Printf("Error saving media playlist: %v", err)
		return err
	}

	if err := os.WriteFile(diskPath, buf.Bytes(), 0644); err != nil {
		log.Printf("Error saving media playlist: %v", err)
		return err
	}

	log.Printf("Saved master playlist: %s", fname)

	for _, segment := range pl.Segments {

		if segment == nil {
			continue
		}

		rel, err := url.Parse(segment.URI)
		if err != nil {
			continue
		}
		absURL := hvi.URL.ResolveReference(rel).String()
		hvi.fetchAndSaveSegment(absURL, filepath.Join(hvi.BaseDir, segment.URI))
	}
	return nil
}

func (hvi *HLSVODIngestor) fetchAndSaveSegment(fullURL, diskPath string) {

	// 1️⃣ Build the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fullURL, nil)
	if err != nil {
		log.Printf("segment request error: %v", err)
		return
	}
	resp, err := hvi.client.Do(req)
	if err != nil {
		log.Printf("segment download error: %v", err)
		return
	}
	defer resp.Body.Close()

	// 2️⃣ Prepare on-disk directories
	if err := os.MkdirAll(filepath.Dir(diskPath), 0755); err != nil {
		log.Printf("mkdir error for %s: %v", diskPath, err)
		return
	}

	// 3️⃣ Write to a temporary file
	tmpPath := diskPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		log.Printf("file create error (%s): %v", tmpPath, err)
		return
	}

	// Ensure closure on all paths
	defer func() {
		tmpFile.Close()
		// If rename never happened, remove the stale tmp file
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	// 4️⃣ Stream download into the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		log.Printf("write error for %s: %v", tmpPath, err)
		return
	}

	// 5️⃣ Flush OS buffers to disk
	if err := tmpFile.Sync(); err != nil {
		log.Printf("sync error for %s: %v", tmpPath, err)
		return
	}
	tmpFile.Close() // close before rename

	// 6️⃣ Atomically move into final place
	if err := os.Rename(tmpPath, diskPath); err != nil {
		log.Printf("rename error from %s to %s: %v", tmpPath, diskPath, err)
	}
}
