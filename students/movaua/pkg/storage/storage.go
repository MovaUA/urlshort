package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

var (
	backetName = []byte("pathsToURLs")
)

// Store reads and persists the data
type Store struct {
	db    *bolt.DB
	close func()
}

// New creates new *Store
func New(filename string) (*Store, error) {
	db, err := bolt.Open(filename, os.FileMode(0600), nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	go stat(ctx, db)

	return &Store{
		db:    db,
		close: cancel,
	}, nil
}

func stat(ctx context.Context, db *bolt.DB) {
	// Grab the initial stats.
	prev := db.Stats()

	t := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-t.C:
			stats := db.Stats()
			diff := stats.Sub(&prev)

			// Encode stats to JSON and print to STDERR.
			json.NewEncoder(os.Stderr).Encode(diff)

			// Save stats for the next loop.
			prev = stats
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

// Close closes the store
func (s *Store) Close() error {
	s.close()
	return s.db.Close()
}

// Put persists path and url
func (s *Store) Put(path, url string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(backetName)
		if err != nil {
			fmt.Println(err)
			return err
		}

		return b.Put([]byte(path), []byte(url))
	})
}

// Get retrieves URL from the storage by the provided path.
// It returns false if the path is not found.
func (s *Store) Get(path string) (url string, found bool) {
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(backetName)
		if b == nil {
			return nil
		}

		urlBytes := b.Get([]byte(path))
		if urlBytes == nil {
			return nil
		}

		url, found = string(urlBytes), true
		return nil
	})

	return url, found
}
