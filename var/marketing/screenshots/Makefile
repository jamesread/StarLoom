# Regenerates marketing screenshots. Capture scripts expect the Vite SPA at
# https://localhost:5173 (see scripts/). Run `make shoot` with the frontend up,
# or `make screenshots` from the repo root (adds the chromedriver shim).
#
# `make manifest` rebuilds screenshot-manifest.yaml from screenshots.ini.
# `make all` currently refreshes the manifest from existing PNGs; full
# throwaway-stack capture (seed + shoot) can be added later.

ROOT := $(abspath ../../..)
HERE := $(abspath .)

DOCS_ASSETS := $(ROOT)/docs/content/assets/marketing
DOCS_IMAGES := login-space.png parent-home-egypt.png child-home-aztecs.png star-chart-greece.png

.PHONY: all shoot manifest publish clean

all: manifest

shoot:
	PYTHONPATH="$(ROOT)/var/marketing/chromedriver-shim$(if $(PYTHONPATH),:$(PYTHONPATH))" \
		repo-helper screenshot --config screenshots.ini
	$(MAKE) manifest

manifest:
	./generate-manifest.py

publish:
	mkdir -p $(DOCS_ASSETS)
	cp $(DOCS_IMAGES) $(DOCS_ASSETS)/

clean:
	rm -rf tmp
