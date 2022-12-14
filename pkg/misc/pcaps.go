package misc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

// GET a single PCAP file from the worker
func FetchPcapFile(client *http.Client, dir string, workerHost string, id string) error {
	u := fmt.Sprintf("http://%s/pcaps/download/%s", workerHost, id)
	log.Debug().Str("url", u).Msg("Doing PCAP request:")
	res, err := client.Get(u)
	if err != nil {
		log.Error().Err(err).Str("url", u).Msg("PCAP request:")
		return err
	}

	contentDisposition := res.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		log.Error().Err(err).Str("content-disposition", contentDisposition).Msg("Parse media type failure:")
		return err
	}
	filename := params["filename"]

	if res.Body != nil {
		defer res.Body.Close()
	}

	filepath := fmt.Sprintf("%s/%s", dir, filename)
	outFile, err := os.Create(filepath)
	if err != nil {
		log.Error().Err(err).Str("file", filepath).Msg("While creating file:")
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		log.Error().Err(err).Str("file", filepath).Msg("Couldn't copy the download file:")
		return err
	}

	return nil
}

type fetchMergePcapRequest struct {
	Query string   `json:"query"`
	Pcaps []string `json:"pcaps"`
}

// GET merged PCAP file from the worker
func FetchMergedPcapFile(client *http.Client, dir string, query string, pcaps []string, workerHost string) error {
	u := fmt.Sprintf("http://%s/pcaps/merge", workerHost)

	var payload fetchMergePcapRequest
	payload.Query = query
	payload.Pcaps = pcaps
	payloadStr, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Str("url", u).Msg("PCAP merge JSON payload marshal error:")
		return err
	}

	log.Debug().Str("url", u).Msg("Doing PCAP merge request:")
	res, err := client.Post(u, "application/json", bytes.NewBuffer(payloadStr))
	if err != nil {
		log.Error().Err(err).Str("url", u).Msg("PCAP merge request:")
		return err
	}

	contentDisposition := res.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		log.Error().Err(err).Str("content-disposition", contentDisposition).Msg("Parse media type failure:")
		return err
	}
	filename := params["filename"]

	if res.Body != nil {
		defer res.Body.Close()
	}

	filepath := fmt.Sprintf("%s/%s", dir, filename)
	outFile, err := os.Create(filepath)
	if err != nil {
		log.Error().Err(err).Str("file", filepath).Msg("While creating file:")
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		log.Error().Err(err).Str("file", filepath).Msg("Couldn't copy the download file:")
		return err
	}

	return nil
}

// GET the name resolution history from the worker
func FetchNameResolutionHistory(client *http.Client, dir string, workerHost string) error {
	u := fmt.Sprintf("http://%s/pcaps/name-resolution-history", workerHost)
	log.Debug().Str("url", u).Msg("Doing name resolution history request:")
	res, err := client.Get(u)
	if err != nil {
		log.Error().Err(err).Str("url", u).Msg("Name resolution history request:")
		return err
	}

	filepath := fmt.Sprintf("%s/name_resolution_history.json", dir)
	outFile, err := os.Create(filepath)
	if err != nil {
		log.Error().Err(err).Str("file", filepath).Msg("While creating file:")
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, res.Body)
	if err != nil {
		log.Error().Err(err).Str("file", filepath).Msg("Couldn't copy the JSON body:")
		return err
	}

	return nil
}
