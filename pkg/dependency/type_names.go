package dependency

type ContainerType string

const (
	EntriesInserter              ContainerType = "EntriesInserter"
	EntriesProvider              ContainerType = "EntriesProvider"
	EntriesSocketStreamer        ContainerType = "EntriesSocketStreamer"
	EntryStreamerSocketConnector ContainerType = "EntryStreamerSocketConnector"
)
