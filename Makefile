# StarApp — build, test, lint, run
# Running make without arguments builds the project and all subdirectories.

.PHONY: all build test lint run clean protocol frontend service integration-test docs migrate migrate-status screenshots

all: build

build: protocol frontend service

protocol:
	$(MAKE) -wC protocol

frontend:
	$(MAKE) -wC frontend

service:
	$(MAKE) -wC service

docs:
	$(MAKE) -wC docs

test: service-test frontend-test

service-test:
	$(MAKE) -wC service test

frontend-test:
	$(MAKE) -wC frontend test

lint: service-lint frontend-lint

service-lint:
	$(MAKE) -wC service lint

frontend-lint:
	$(MAKE) -wC frontend lint

run:
	$(MAKE) -wC service run

integration-test:
	$(MAKE) -wC integration-tests

migrate:
	DB_PATH="$(DB_PATH)" $(MAKE) -wC database/sqlite

migrate-status:
	DB_PATH="$(DB_PATH)" $(MAKE) -wC database/sqlite status

# Marketing PNGs in var/marketing/. Requires the SPA at https://localhost:5173.
# The shim skips webdriver_manager so Selenium uses the system chromedriver.
screenshots:
	PYTHONPATH="$(CURDIR)/var/marketing/chromedriver-shim$(if $(PYTHONPATH),:$(PYTHONPATH))" \
		repo-helper screenshot

clean:
	$(MAKE) -wC protocol clean
	$(MAKE) -wC frontend clean
	$(MAKE) -wC service clean
	$(MAKE) -wC integration-tests clean
	$(MAKE) -wC docs clean
