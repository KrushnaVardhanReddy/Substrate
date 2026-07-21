#!/bin/bash
docker pull postgres:15-alpine
docker run --name substrate-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=substrate -p 5432:5432 -d postgres:15-alpine
