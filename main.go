package main

import (
	"github.com/sayze/friendly-api/internal/http"
	"github.com/sayze/friendly-api/internal/store"
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
