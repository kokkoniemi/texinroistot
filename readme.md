# Project

* This is a comic catalog project for the villains in Tex Willer series.

## Startup

Copy .env-example to .env and populate values

```
docker-compose up
psql -h localhost -p 5432 -d tex -U tex < texinroistot-server/internal/db/schema.sql
docker-compose exec backend go run cmd/importer/importer.go
```
