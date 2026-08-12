env "local" {
  dev = "docker+mysql://_/mysql:8/dev"
  src = "ent://internal/data/ent/schema"
  url = "mysql://root@localhost:13306/mars"
  migration {
    dir = "file://internal/data/ent/migrate/migrations"
  }
}
