package main

import (
	"github.com/sayze/friendly-api/internal"
	"github.com/sayze/friendly-api/internal/http"
	"github.com/sayze/friendly-api/internal/memory"
)

func main() {

	db := memory.NewService()

	handler, err := http.New(db)

	if err != nil {
		panic(err)
	}

	config := internal.NewConfiguration()

	handler.ListenAndServe(config.Http.Host, config.Http.Port)

}
