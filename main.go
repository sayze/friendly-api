package main

import (
	"github.com/sayze/foundu-taker-api/internal/server"
	"github.com/sayze/foundu-taker-api/internal/store/memory"
)

func main() {
	// Create a memory store of friend entity.
	friendStore, _ := memory.New()

	s, err := server.New(friendStore)

	if err != nil {
		panic(err)
	}

	s.ListenAndServe()

}
