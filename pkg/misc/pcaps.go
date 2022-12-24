package misc

import (
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

// GET merged PCAP file from the worker
func FetchMergedPcapFile(client *http.Client, dir string, workerHost string) error {
	u := fmt.Sprintf("http://%s/pcaps/merge", workerHost)
	log.Debug().Str("url", u).Msg("Doing PCAP merge request:")
	res, err := client.Get(u)
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

	filepath := fmt.Sprintf("%s/worker_%s_%s", dir, workerHost, filename)
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

	filepath := fmt.Sprintf("%s/worker_%s_name_resolution_history.json", dir, workerHost)
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
