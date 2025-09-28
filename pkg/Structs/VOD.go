package Structs

type VODIngestRequest struct {
	Name    string `json:"name"`
	Episode string `json:"episode"`
	Url     string `json:"url"`
}
