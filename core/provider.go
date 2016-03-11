package core

type Provider interface {
	GetBulkSize() int
	Publish(JournalEntryIterator) (int, error)
}