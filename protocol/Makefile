.PHONY: all buf clean

all: buf

buf:
	@if ! find . -name '*.proto' -print -quit | grep -q .; then \
		echo "no .proto files yet; skipping code generation"; \
	elif command -v buf >/dev/null 2>&1; then \
		buf generate; \
	else \
		echo "buf not installed; skipping code generation"; \
	fi

clean:
	rm -rf ../service/gen
