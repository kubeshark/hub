package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	baseApi "github.com/kubeshark/base/pkg/api"
	"github.com/kubeshark/hub/pkg/dependency"
	"github.com/kubeshark/hub/pkg/har"
	"github.com/kubeshark/hub/pkg/holder"
	"github.com/kubeshark/hub/pkg/oas"
	"github.com/kubeshark/hub/pkg/providers"
	"github.com/kubeshark/hub/pkg/resolver"
	"github.com/kubeshark/hub/pkg/servicemap"
	"github.com/kubeshark/hub/pkg/utils"
	"github.com/rs/zerolog/log"
)

var k8sResolver *resolver.Resolver

func StartResolving(namespace string) {
	errOut := make(chan error, 100)
	res, err := resolver.NewFromInCluster(errOut, namespace)
	if err != nil {
		log.Error().Err(err).Msg("While creating K8s resolver!")
		return
	}
	ctx := context.Background()
	res.Start(ctx)
	go func() {
		for {
			err := <-errOut
			log.Error().Err(err).Msg("Name resolution failed:")
		}
	}()

	k8sResolver = res
	holder.SetResolver(res)
}

func StartReadingEntries(harChannel <-chan *baseApi.OutputChannelItem, workingDir *string, extensionsMap map[string]*baseApi.Extension) {
	if workingDir != nil && *workingDir != "" {
		startReadingFiles(*workingDir)
	} else {
		startReadingChannel(harChannel, extensionsMap)
	}
}

func startReadingFiles(workingDir string) {
	if err := os.MkdirAll(workingDir, os.ModePerm); err != nil {
		log.Error().Err(err).Str("dir", workingDir).Msg("Failed to create directory!")
		return
	}

	for {
		dir, _ := os.Open(workingDir)
		dirFiles, _ := dir.Readdir(-1)

		var harFiles []os.FileInfo
		for _, fileInfo := range dirFiles {
			if strings.HasSuffix(fileInfo.Name(), ".har") {
				harFiles = append(harFiles, fileInfo)
			}
		}
		sort.Sort(utils.ByModTime(harFiles))

		if len(harFiles) == 0 {
			log.Info().Msg("Waiting for new files")
			time.Sleep(3 * time.Second)
			continue
		}
		fileInfo := harFiles[0]
		inputFilePath := path.Join(workingDir, fileInfo.Name())
		file, err := os.Open(inputFilePath)
		utils.CheckErr(err)

		var inputHar har.HAR
		decErr := json.NewDecoder(bufio.NewReader(file)).Decode(&inputHar)
		utils.CheckErr(decErr)

		rmErr := os.Remove(inputFilePath)
		utils.CheckErr(rmErr)
	}
}

func startReadingChannel(outputItems <-chan *baseApi.OutputChannelItem, extensionsMap map[string]*baseApi.Extension) {
	for item := range outputItems {
		extension := extensionsMap[item.Protocol.Name]
		resolvedSource, resolvedDestination, namespace := resolveIP(item.ConnectionInfo)

		if namespace == "" && item.Namespace != baseApi.UnknownNamespace {
			namespace = item.Namespace
		}

		kubesharkEntry := extension.Dissector.Analyze(item, resolvedSource, resolvedDestination, namespace)

		data, err := json.Marshal(kubesharkEntry)
		if err != nil {
			log.Error().Err(err).Msg("While marshaling entry!")
			continue
		}

		entryInserter := dependency.GetInstance(dependency.EntriesInserter).(EntryInserter)
		if err := entryInserter.Insert(kubesharkEntry); err != nil {
			log.Error().Err(err).Msg("While inserting entry!")
		}

		summary := extension.Dissector.Summarize(kubesharkEntry)
		providers.EntryAdded(len(data), summary)

		serviceMapGenerator := dependency.GetInstance(dependency.ServiceMapGeneratorDependency).(servicemap.ServiceMapSink)
		serviceMapGenerator.NewTCPEntry(kubesharkEntry.Source, kubesharkEntry.Destination, &item.Protocol)

		oasGenerator := dependency.GetInstance(dependency.OasGeneratorDependency).(oas.OasGeneratorSink)
		oasGenerator.HandleEntry(kubesharkEntry)
	}
}

func resolveIP(connectionInfo *baseApi.ConnectionInfo) (resolvedSource string, resolvedDestination string, namespace string) {
	if k8sResolver != nil {
		unresolvedSource := connectionInfo.ClientIP
		resolvedSourceObject := k8sResolver.Resolve(unresolvedSource)
		if resolvedSourceObject == nil {
			log.Debug().Str("source", unresolvedSource).Msg("Cannot find resolved name!")
			if os.Getenv("SKIP_NOT_RESOLVED_SOURCE") == "1" {
				return
			}
		} else {
			resolvedSource = resolvedSourceObject.FullAddress
			namespace = resolvedSourceObject.Namespace
		}

		unresolvedDestination := fmt.Sprintf("%s:%s", connectionInfo.ServerIP, connectionInfo.ServerPort)
		resolvedDestinationObject := k8sResolver.Resolve(unresolvedDestination)
		if resolvedDestinationObject == nil {
			log.Debug().Str("destination", unresolvedDestination).Msg("Cannot find resolved name!")
			if os.Getenv("SKIP_NOT_RESOLVED_DEST") == "1" {
				return
			}
		} else {
			resolvedDestination = resolvedDestinationObject.FullAddress
			// Overwrite namespace (if it was set according to the source)
			// Only overwrite if non-empty
			if resolvedDestinationObject.Namespace != "" {
				namespace = resolvedDestinationObject.Namespace
			}
		}
	}
	return resolvedSource, resolvedDestination, namespace
}

func CheckIsServiceIP(address string) bool {
	if k8sResolver == nil {
		return false
	}
	return k8sResolver.CheckIsServiceIP(address)
}
