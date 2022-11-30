package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/kubeshark/hub/pkg/dependency"
	"github.com/kubeshark/hub/pkg/oas"
	"github.com/kubeshark/hub/pkg/servicemap"

	"github.com/kubeshark/hub/pkg/har"
	"github.com/kubeshark/hub/pkg/holder"
	"github.com/kubeshark/hub/pkg/providers"

	"github.com/kubeshark/hub/pkg/resolver"
	"github.com/kubeshark/hub/pkg/utils"

	baseApi "github.com/kubeshark/base/pkg/api"
)

var k8sResolver *resolver.Resolver

func StartResolving(namespace string) {
	errOut := make(chan error, 100)
	res, err := resolver.NewFromInCluster(errOut, namespace)
	if err != nil {
		log.Printf("error creating k8s resolver %s", err)
		return
	}
	ctx := context.Background()
	res.Start(ctx)
	go func() {
		for {
			err := <-errOut
			log.Printf("name resolving error %s", err)
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
		log.Printf("Failed to make dir: %s, err: %v", workingDir, err)
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
			log.Printf("Waiting for new files")
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
	if outputItems == nil {
		panic("Channel of captured messages is nil")
	}

	for item := range outputItems {
		extension := extensionsMap[item.Protocol.Name]
		resolvedSource, resolvedDestination, namespace := resolveIP(item.ConnectionInfo)

		if namespace == "" && item.Namespace != baseApi.UnknownNamespace {
			namespace = item.Namespace
		}

		kubesharkEntry := extension.Dissector.Analyze(item, resolvedSource, resolvedDestination, namespace)

		data, err := json.Marshal(kubesharkEntry)
		if err != nil {
			log.Printf("Error while marshaling entry: %v", err)
			continue
		}

		entryInserter := dependency.GetInstance(dependency.EntriesInserter).(EntryInserter)
		if err := entryInserter.Insert(kubesharkEntry); err != nil {
			log.Printf("Error inserting entry, err: %v", err)
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
			log.Printf("Cannot find resolved name to source: %s", unresolvedSource)
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
			log.Printf("Cannot find resolved name to dest: %s", unresolvedDestination)
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
