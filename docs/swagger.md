# Swagger scaffold

В проекте подготовлен каркас для генерации Swagger через `swaggo/swag`.

## Как сгенерировать локально

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go -o docs/swagger
```

После этого можно подключить `http-swagger` или отдавать сгенерированные файлы через отдельный route.

Из-за ограничений среды генерация не выполнялась автоматически, поэтому этот пункт оставлен как готовый scaffold.
