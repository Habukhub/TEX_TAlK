data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./src/atlas/loader",
  ]
}

env "local" {
  src = data.external_schema.gorm.url
  url = "postgres://admin:password@localhost:6632/tex_db?sslmode=disable"
  dev = "docker://postgres/18/dev?search_path=public"

  migration {
    dir = "file://migrations/local"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "prod" {
  src = data.external_schema.gorm.url
  url = getenv("DATABASE_URL")

  migration {
    dir = "file://migrations/prod"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
