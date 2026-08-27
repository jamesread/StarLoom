.PHONY: all build serve clean deps

all: build

deps:
	pip install -r requirements.txt

build: deps
	mkdocs build -f mkdocs.yml

serve: deps
	mkdocs serve -f mkdocs.yml

clean:
	rm -rf site
