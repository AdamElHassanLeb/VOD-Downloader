package Controllers

import (
	"encoding/json"
	"github.com/AdamElHassanLeb/VOD-Downloader/API/pkg/Structs"
	"net/http"
)

func IngestVOD(w http.ResponseWriter, r *http.Request) {

	var ingestRequest Structs.VODIngestRequest

	if err := json.NewDecoder(r.Body).Decode(&ingestRequest); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := Service.VODIngest.IngestHLS(ingestRequest); err != nil {
		http.Error(w, "An error has occurred: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}
