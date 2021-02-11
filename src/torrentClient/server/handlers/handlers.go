package handlers

import (
	"bytes"
	"net/http"

	"torrent_client/db"
	"torrent_client/torrentfile"

	"github.com/sirupsen/logrus"
)

func DownloadRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fileId := r.URL.Query().Get("file_id")

		response := struct {
			IsLoaded	bool	`json:"isLoaded"`
			Key			string	`json:"key"`
		}{}

		response.IsLoaded = true
		response.Key = fileId

		if db.GetLoadedStateDb().CheckIfFileIsActiveLoading(fileId) {
			response.IsLoaded = true
		} else {
			response.IsLoaded = false
		}

		torrentBytes := db.GetFilesManagerDb().GetTorrentFileForByFileId(fileId)
		if torrentBytes == nil {
			SendFailResponseWithCode(w, "File not found", http.StatusNotFound)
			return
		}

		go func() {
			torrent, err := torrentfile.GetManager().ReadTorrentFileFromHttpBody(bytes.NewBuffer(torrentBytes))
			if err != nil {
				logrus.Errorf("Error reading torrent file: %v", err)
				SendFailResponseWithCode(w, "Error reading body: " + err.Error(), http.StatusInternalServerError)
			}

			torrent.FileId = fileId

			err = torrent.DownloadToFile()
			if err != nil {
				logrus.Errorf("Error downloading to file: %v", err)
			}
		}()

		SendDataResponse(w, response)
	}
}
