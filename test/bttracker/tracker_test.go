package torrent

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

type TrackerRequestData struct {
	peerID     string
	infoHash   string
	event      string
	downloaded uint64
	uploaded   uint64
	remaining  uint64
}

func sendHTTPRequest(data *TrackerRequestData) (string, error) {

	url := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&event=%s&left=%d&compact=0&downloaded=%d&uploaded=%d&port=6881",
		"http://127.0.0.1:8080/announce",
		url.QueryEscape(data.infoHash),
		url.QueryEscape(data.peerID),
		data.event, //completed,started,stopped
		data.remaining,
		data.downloaded,
		data.uploaded,
	)
	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad response from tracker (%d): %s",
			resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
	/*
		resp, err := bencode.Decode(httpResponse.Body)
		if err != nil {
			return err
		}

		if failureReason, exists := resp["failure reason"]; exists {
			return errors.New(failureReason.(string))
		}

		interval := resp["interval"].(int64)
		peers := resp["peers"]
		return nil
	*/
}

func Test_Tracker_Started(t *testing.T) {
	data := &TrackerRequestData{
		peerID:     "peerid111111",
		infoHash:   "infohash111111",
		event:      "started",
		downloaded: 900,
		uploaded:   800,
		remaining:  500,
	}

	if resp, err := sendHTTPRequest(data); err != nil {
		t.Error(err)
	} else {
		t.Log(resp)
	}

	data = &TrackerRequestData{
		peerID:     "peerid222222",
		infoHash:   "infohash111111",
		event:      "started",
		downloaded: 900,
		uploaded:   800,
		remaining:  500,
	}

	if resp, err := sendHTTPRequest(data); err != nil {
		t.Error(err)
	} else {
		t.Log(resp)
	}

}

func Test_Tracker_Completed(t *testing.T) {

}

func Test_Tracker_Stopped(t *testing.T) {

}
