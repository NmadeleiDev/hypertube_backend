package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"hypertube_storage/model"
	"hypertube_storage/parser/env"

	"github.com/sirupsen/logrus"
)

func SendFailResponseWithCode(w http.ResponseWriter, text string, code int) {
	var packet []byte
	var err error

	response := &model.DataResponse{Status: false, Data: text}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	if packet, err = json.Marshal(response); err != nil {
		logrus.Error("Error marshalling response: ", err)
	}
	if _, err = w.Write(packet); err != nil {
		logrus.Error("Error sending response: ", err)
	}
}

func GetContentTypeForReqType(reqType string) string {
	switch reqType {
	case videoRequest:
		return "video/mp4"
	case subtitlesRequest:
		return "text/vtt"
	default:
		return "plain/text"
	}
}

func GetResponseStatusForReqType(reqType string) int {
	switch reqType {
	case videoRequest:
		return http.StatusPartialContent
	case subtitlesRequest:
		return http.StatusOK
	default:
		return http.StatusOK
	}
}

func SendTaskToTorrentClient(fileId string) bool {
	req, err := http.Get(fmt.Sprintf("http://%s/download/%s", env.GetParser().GetLoaderServiceHost(), fileId))
	if err != nil {
		logrus.Errorf("Error calling loader service: %v", err)
		return false
	}

	if req.StatusCode != http.StatusOK {
		logrus.Errorf("Not ok status from torrent client: %v %v", req.StatusCode, req.Status)
		return false
	}

	info := model.LoaderTaskResponse{}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logrus.Errorf("Error reading body: %v", err)
		return false
	}

	if err := json.Unmarshal(body, &info); err != nil {
		logrus.Errorf("Error unmarshal body from loader: %v", err)
		return false
	}

	return true
}


