env "local" {
  // dev 库的 MySQL 大版本与 CI（.github/workflows/test.yaml）和 dev/docker-compose.yml
  // 统一为 8.4 LTS。`mysql:8` 浮 tag 实际解析到 8.0.x（已 EOL），显式写 8.4 才能锁住大版本。
  dev = "docker+mysql://_/mysql:8.4/dev"
  src = "ent://internal/data/ent/schema"
  url = "mysql://root@localhost:13306/mars"
  migration {
    dir = "file://internal/data/ent/migrate/migrations"
  }
}
