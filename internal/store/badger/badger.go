package badger

import (
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

type BadgerStore struct {
	db               *badger.DB
	garbageCollector *time.Ticker
}

func NewBadgerSTore(dbFilePath string) (*BadgerStore, error) {
	db, err := badger.Open(badger.DefaultOptions(dbFilePath))

	gc := time.NewTicker(60 * time.Second)

	store := &BadgerStore{
		db:               db,
		garbageCollector: gc,
	}

	go func() {
		for {
			<-store.garbageCollector.C
			store.GarbageCollector()
		}
	}()

	return store, err
}

func (store *BadgerStore) Close() error {
	store.garbageCollector.Stop()
	return store.db.Close()
}

func (store *BadgerStore) GarbageCollector() error {
	// Per Badger's docs - run the garbage collection
	// until it returns an error
again:
	err := store.db.RunValueLogGC(0.7)
	if err == nil {
		goto again
	}
	return store.db.RunValueLogGC(0.5)
}

func (store *BadgerStore) Migrate() error {
	return nil
}
