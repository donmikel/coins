.PHONY: help
help:
	@echo "usage: make [target]"
	@echo
	@echo "targets:"
	@echo "  help            show this message"
	@echo "  test            run all tests (requires docker)"
	@echo "  clean-docker    stop test docker containers"
	@echo

.PHONY: test
test:
	# Starting PostgesSQL docker container ...
	@docker run --rm -d --name test_pg -p 54322:5432 -e POSTGRES_PASSWORD=test -e POSTGRES_USER=test -v my_pgvolume1:/var/lib/postgresql/data postgres:12 > /dev/null
	@echo "SELECT 'CREATE DATABASE coins' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'coins')\gexec" | docker exec -i test_pg psql -U test
	@docker exec -i test_pg psql -U test coins < pg/script.sql

	# Waiting 3s before running tests...
	@sleep 3sd
	# Running tests ...
	@\
		TEST_POSTGRES_ADDRESS='localhost' \
		TEST_POSTGRES_PASSWORD='test' \
		TEST_POSTGRES_USER='test' \
		TEST_POSTGRES_DATABASE='coins' \
		TEST_POSTGRES_PORT=54322 \
		go test -mod=vendor -cover -race -count=1 ./...; \
		rc=$$?; \
		docker stop test_pg > /dev/null; \
		exit $$rc
	# OK

.PHONY: clean-docker
clean-docker:
	@docker stop test_pg >/dev/null 2>&1 || true