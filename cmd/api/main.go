// Command api is friendly-api's entrypoint and composition root.
package main

import (
	"log"
	"net/http"

	"github.com/sayze/friendly-api/internal/config"
	"github.com/sayze/friendly-api/internal/friend/infra/cloudinary"
	"github.com/sayze/friendly-api/internal/friend/infra/memory"
	"github.com/sayze/friendly-api/internal/friend/service"
	"github.com/sayze/friendly-api/internal/server"
)

func main() {
	cfg := config.Load()

	repo := memory.NewRepository()
	images := cloudinary.NewImageStore(cloudinary.Config{
		UploadURL: cfg.Cdn.UploadURL,
		APIKey:    cfg.Cdn.APIKey,
		APISecret: cfg.Cdn.APISecret,
	})
	svc := service.NewService(repo, images)

	srv := server.New(svc)

	log.Printf("friendly-api listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, srv); err != nil {
		log.Fatal(err)
	}
}
