package main

import (
	"github.com/sayze/friendly-api/database/memory"
	"github.com/sayze/friendly-api/http"
)

func main() {

	db := memory.NewService()

	handler, err := http.NewHandler(db)

	if err != nil {
		panic(err)
	}

	config := NewConfiguration()

	handler.ListenAndServe(config.Http.Host, config.Http.Port)

}
