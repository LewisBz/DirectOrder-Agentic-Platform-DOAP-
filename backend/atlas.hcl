env "local" {
  src = "file://migrations/schema.sql"
  url = getenv("OWNER_DATABASE_URL")
  dev = "docker://postgres/16/dev"
}
