COMPOSE_FILE := $(realpath deploy/podman/compose.yaml)

deploy-db:
	@echo "🚀 Deploy DB on podman..."
	podman rm -f placement-db || true
	podman volume rm podman_placement-db || true
	podman-compose -f $(COMPOSE_FILE) up -d placement-db
	test/scripts/wait_for_postgres.sh podman
	podman exec -it placement-db psql -c 'ALTER ROLE admin WITH SUPERUSER'
	podman exec -it placement-db createdb admin || true
	@echo "✅ DB was deployed successfully on podman."

kill-db:
	@echo "🗑️ Remove DB instance from podman..."
	podman-compose -f $(COMPOSE_FILE) down placement-db
	@echo "✅ DB instance was removed successfully from podman."

.PHONY: deploy-db kill-db
