package store_test

import (
	store "memcached/internal"
	"testing"
)

func TestStore(t *testing.T) {
	store := store.NewInMemoryStore()
	t.Run("add new key to the store", func(t *testing.T) {
		key := "test"
		data := []byte("some piece of data")
		var flags uint16 = 1

		err := store.Add(key, data, flags)

		if err != nil {
			t.Errorf("an error occurred while storing data: %s", err)
		}
	})
	t.Run("get key from the store", func(t *testing.T) {
		key := "test"
		expectedData := "some piece of data"

		entry, err := store.Get(key)

		if err != nil {
			t.Errorf("an error occurred while fetching key from store: %s", err)
		}

		if string(entry.Data) != expectedData {
			t.Errorf("expected key to store %s, found %s", expectedData, string(entry.Data))
		}
	})
}
