package badger

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/hlfshell/coppermind/pkg/chat"
)

func (store *BadgerStore) insertMessageEntries(msg *chat.Message) ([]*badger.Entry, error) {
	/*
		Inserting a message results in multiple keys being updated.

		1. The message is saved under its ID

		2. The message is added as a part of the conversation/messages set (or
			created if it doesn't exist)

		3. The conversation latest time is updated with this message's creation
			time.
	*/

	// Message insert
	payload, err := msg.JSON()
	if err != nil {
		return nil, err
	}
	msgEntry := badger.NewEntry([]byte(msg.ID), []byte(payload))

	// Conversation insert
	//TODO - can I just append repeatedly?
	convEntry := badger.NewEntry(
		[]byte(msg.Conversation), []byte(msg.ID),
	)

	return []*badger.Entry{msgEntry, convEntry}, nil
}

func (store *BadgerStore) SaveMessage(msg *chat.Message) error {
	return store.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(
			[]byte("answer"),
			[]byte("42"),
		)
		err := txn.SetEntry(e)
		return err
	})
}
