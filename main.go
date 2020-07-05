package main

import (
	"github.com/sayze/friendly-api/internal/http"
	"github.com/sayze/friendly-api/internal/memory"
	"github.com/sayze/friendly-api/internal/store"
)

func main() {

	db := memory.NewService()

	handler, err := http.New(db)

	if err != nil {
		panic(err)
	}

	handler.ListenAndServe()

}
