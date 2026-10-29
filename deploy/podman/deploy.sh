#!/bin/bash
set -e

MIRROR="15dczmo6htbpugrx1i.xuanyuan.run"
PG_IMAGE="${MIRROR}/library/postgres:17-alpine"
REDIS_IMAGE="${MIRROR}/library/redis:8.8.4-trixie"
POD_NAME="dcs-pod"

echo "==> Pulling images..."
podman pull "$PG_IMAGE"
podman pull "$REDIS_IMAGE"

echo "==> Creating pod..."
podman pod exists "$POD_NAME" 2>/dev/null || podman pod create --name "$POD_NAME" -p 5432:5432 -p 6379:6379

echo "==> Starting PostgreSQL..."
podman rm -f dcs-postgres 2>/dev/null || true
podman volume exists dcs-pgdata 2>/dev/null || podman volume create dcs-pgdata
podman run -d --pod "$POD_NAME" --name dcs-postgres \
  -e POSTGRES_USER=dcs \
  -e POSTGRES_PASSWORD=dcs123 \
  -e POSTGRES_DB=dcs \
  -v dcs-pgdata:/var/lib/postgresql/data \
  "$PG_IMAGE"

echo "==> Starting Redis..."
podman rm -f dcs-redis 2>/dev/null || true
podman volume exists dcs-redisdata 2>/dev/null || podman volume create dcs-redisdata
podman run -d --pod "$POD_NAME" --name dcs-redis \
  -v dcs-redisdata:/data \
  "$REDIS_IMAGE" redis-server --appendonly yes

echo "==> Waiting for services..."
sleep 3
podman exec dcs-postgres psql -U dcs -d dcs -c "SELECT 1;" > /dev/null 2>&1 && echo "    PostgreSQL: OK" || echo "    PostgreSQL: FAIL"
podman exec dcs-redis redis-cli ping 2>/dev/null | grep -q PONG && echo "    Redis: OK" || echo "    Redis: FAIL"

echo "==> Done!"
echo "    PostgreSQL: localhost:5432 (user=dcs, password=dcs123, db=dcs)"
echo "    Redis:      localhost:6379"
