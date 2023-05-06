package store

type Store interface {
	LowLevelStore
	HighLevelStore
}
