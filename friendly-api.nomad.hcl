variable "image" {
  type    = string
  default = "sayze/friendly-api:master"
}

variable "domain" {
  type    = string
  default = "friendly-api.sayedsadeed.com"
}

variable "cdn_upload_url" {
  type    = string
  default = "https://api.cloudinary.com/v1_1/sayze/image/upload"
}

job "friendly-api" {
  datacenters = ["hl"]
  type        = "service"

  group "friendly-api" {
    count = 1

    network {
      port "http" {
        to = 4040
      }
    }

    service {
      name     = "friendly-api"
      port     = "http"
      provider = "consul"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.friendly-api.rule=Host(`${var.domain}`)",
        "traefik.http.routers.friendly-api.entrypoints=websecure",
        "traefik.http.routers.friendly-api.tls=true",
      ]

      check {
        type     = "http"
        path     = "/"
        interval = "30s"
        timeout  = "5s"
      }
    }

    task "friendly-api" {
      driver = "docker"

      vault {
        policies = ["nomad"]
      }

      config {
        image = var.image
        ports = ["http"]
      }

      env {
        ADDR           = ":4040"
        CDN_UPLOAD_URL = var.cdn_upload_url
      }

      template {
        data        = <<-EOF
          {{ with secret "secret/data/homelab/friendly-api" }}
          CDN_API_KEY="{{ .Data.data.cdn_api_key }}"
          CDN_API_SECRET="{{ .Data.data.cdn_api_secret }}"
          {{ end }}
        EOF
        destination = "secrets/env"
        env         = true
      }

      logs {
        max_files     = 3
        max_file_size = 10
      }

      resources {
        cpu    = 50
        memory = 64
      }
    }
  }
}
