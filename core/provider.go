package core

type ProviderConfig interface {
	Name() string
	BulkSize() int
}

type Provider interface {
	Publish(JournalEntryIterator) (int, error)
}