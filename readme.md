# Project

* This is a comic catalog project for the villains in Tex Willer series.

## Startup

Copy .env-example to .env and populate values

```
docker-compose up
psql -h localhost -p 5432 -d tex -U tex < texinroistot-server/internal/db/schema.sql
```

Helper script for import + activate newest version:

```
./scripts/import_excel_and_activate_latest.sh
```
