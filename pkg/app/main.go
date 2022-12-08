package app

import (
	"sort"

	baseApi "github.com/kubeshark/base/pkg/api"
	amqpExt "github.com/kubeshark/base/pkg/extensions/amqp"
	httpExt "github.com/kubeshark/base/pkg/extensions/http"
	kafkaExt "github.com/kubeshark/base/pkg/extensions/kafka"
	redisExt "github.com/kubeshark/base/pkg/extensions/redis"
	"github.com/kubeshark/hub/pkg/api"
	"github.com/kubeshark/hub/pkg/providers"
)

var (
	Extensions    []*baseApi.Extension          // global
	ExtensionsMap map[string]*baseApi.Extension // global
	ProtocolsMap  map[string]*baseApi.Protocol  //global
)

func LoadExtensions() {
	Extensions = make([]*baseApi.Extension, 0)
	ExtensionsMap = make(map[string]*baseApi.Extension)
	ProtocolsMap = make(map[string]*baseApi.Protocol)

	extensionHttp := &baseApi.Extension{}
	dissectorHttp := httpExt.NewDissector()
	dissectorHttp.Register(extensionHttp)
	extensionHttp.Dissector = dissectorHttp
	Extensions = append(Extensions, extensionHttp)
	ExtensionsMap[extensionHttp.Protocol.Name] = extensionHttp
	protocolsHttp := dissectorHttp.GetProtocols()
	for k, v := range protocolsHttp {
		ProtocolsMap[k] = v
	}

	extensionAmqp := &baseApi.Extension{}
	dissectorAmqp := amqpExt.NewDissector()
	dissectorAmqp.Register(extensionAmqp)
	extensionAmqp.Dissector = dissectorAmqp
	Extensions = append(Extensions, extensionAmqp)
	ExtensionsMap[extensionAmqp.Protocol.Name] = extensionAmqp
	protocolsAmqp := dissectorAmqp.GetProtocols()
	for k, v := range protocolsAmqp {
		ProtocolsMap[k] = v
	}

	extensionKafka := &baseApi.Extension{}
	dissectorKafka := kafkaExt.NewDissector()
	dissectorKafka.Register(extensionKafka)
	extensionKafka.Dissector = dissectorKafka
	Extensions = append(Extensions, extensionKafka)
	ExtensionsMap[extensionKafka.Protocol.Name] = extensionKafka
	protocolsKafka := dissectorKafka.GetProtocols()
	for k, v := range protocolsKafka {
		ProtocolsMap[k] = v
	}

	extensionRedis := &baseApi.Extension{}
	dissectorRedis := redisExt.NewDissector()
	dissectorRedis.Register(extensionRedis)
	extensionRedis.Dissector = dissectorRedis
	Extensions = append(Extensions, extensionRedis)
	ExtensionsMap[extensionRedis.Protocol.Name] = extensionRedis
	protocolsRedis := dissectorRedis.GetProtocols()
	for k, v := range protocolsRedis {
		ProtocolsMap[k] = v
	}

	sort.Slice(Extensions, func(i, j int) bool {
		return Extensions[i].Protocol.Priority < Extensions[j].Protocol.Priority
	})

	api.InitMaps(ExtensionsMap, ProtocolsMap)
	providers.InitProtocolToColor(ProtocolsMap)
}

func GetEntryInputChannel() chan *baseApi.OutputChannelItem {
	outputItemsChannel := make(chan *baseApi.OutputChannelItem)
	filteredOutputItemsChannel := make(chan *baseApi.OutputChannelItem)
	go FilterItems(outputItemsChannel, filteredOutputItemsChannel)
	go api.StartReadingEntries(filteredOutputItemsChannel, nil, ExtensionsMap)

	return outputItemsChannel
}

func FilterItems(inChannel <-chan *baseApi.OutputChannelItem, outChannel chan *baseApi.OutputChannelItem) {
	for message := range inChannel {
		if message.ConnectionInfo.IsOutgoing && api.CheckIsServiceIP(message.ConnectionInfo.ServerIP) {
			continue
		}

		outChannel <- message
	}
}
