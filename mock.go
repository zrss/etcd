package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("frag.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return err
		}
		return err
	})

	go func() {
		db.View(func(tx *bolt.Tx) error {
			fmt.Printf("start of long run read txn\n")
			fmt.Printf("read txn txid: %d\n", tx.ID())
			bucket := tx.Bucket([]byte("MyBucket"))
			bucket.Get([]byte("answer"))

			<-time.After(10 * time.Second)
			fmt.Printf("end of long run read txn\n")
			return nil
		})
	}()

	mockValue := make([]byte, 1024)
	for i := 0; i < 64; i++ {
		db.Update(func(tx *bolt.Tx) error {
			fmt.Printf("rw txn txid: %d\n", tx.ID())
			b := tx.Bucket([]byte("MyBucket"))
			err = b.Put([]byte("answer"), mockValue)
			return err
		})
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()
}
