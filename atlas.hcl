env "local" {
  dev = "docker+mysql://_/mysql:8/dev"
  src = "ent://internal/ent/schema"
  url = "mysql://root:root@localhost:13306/mars"
  migration {
    dir = "file://internal/ent/migrate/migrations"
  }
}