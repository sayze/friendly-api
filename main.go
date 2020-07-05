package main

import (
	"github.com/sayze/foundu-taker-api/internal/http"
	"github.com/sayze/foundu-taker-api/internal/store"
)

func main() {
	// Create a memory store of friend entity.
	friendStore, _ := store.New()

	s, err := http.New(friendStore)

	if err != nil {
		panic(err)
	}

	s.ListenAndServe()

}
