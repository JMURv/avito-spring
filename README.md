[![Go Coverage](https://github.com/JMURv/avito-spring/wiki/coverage.svg)](https://raw.githack.com/wiki/JMURv/avito-spring/coverage.html)

## Запуск проекта
### Конфигурация
Скопировать `example.config.yaml` как `config.yaml`:
```sh
cp configs/example.config.yaml configs/config.yaml
```

Перейти в папку build:
```sh
cd build
```

Скопировать `.env.example` как `.env`:
```sh
cp compose/env/.env.example compose/env/.env
```

Запустить docker-compose:
```sh
docker compose --env-file compose/env/.env -f compose/dc.yaml up --build
```
Вместе с приложением и базой также запустятся контейнеры для `prometheus`, `node-exporter`

| Сервис        | Адрес                 |
|---------------|-----------------------|
| App           | http://localhost:8080 |
| App (GRPC)    | http://localhost:3000 |
| App (Metrics) | http://localhost:9000 |
| DB            | http://localhost:5432 |
| Prometheus    | http://localhost:9090 |
| Node Exporter | http://localhost:9100 |

### Запуск интеграционного теста
```sh
cd build
```

Находясь в папке `build`:
```sh
docker compose --env-file compose/env/.env.test -f compose/dc.test.yaml up -d
```

Указываем путь до миграций через переменную окружения:
```sh
export MIGRATIONS_PATH=../../../internal/repo/db/migration
```

Возвращаемся в корень проекта:
```sh
cd ..
```

Запускаем тест
```sh
go test -v ./tests/integration/...
```

