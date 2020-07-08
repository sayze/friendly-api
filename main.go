package main

import (
	"github.com/sayze/friendly-api/internal"
	"github.com/sayze/friendly-api/internal/database/memory"
	"github.com/sayze/friendly-api/internal/http"
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
