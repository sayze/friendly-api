package main

import (
	"github.com/sayze/friendly-api/config"
	"github.com/sayze/friendly-api/database/memory"
	"github.com/sayze/friendly-api/http"
)

func main() {

	db := memory.NewService()

	handler, err := http.NewHandler(db)

	if err != nil {
		panic(err)
	}

	conf := config.NewConfiguration()

	handler.ListenAndServe(conf.Http.Host, conf.Http.Port)

}
