HEAD = `git rev-parse HEAD`
TIME = `date +%FT%T%z`
BINARY = "admin_launcher"


default: gotool
	@CGO_ENABLED=0 go build -ldflags "-X main.Version=${HEAD} -X main.BuildTime=${TIME}" -o ${BINARY}

clean:
	rm -f ${binary}
	find . -name "[._]*.s[a-w][a-z]" | xargs -i rm -f {}

gotool:
	@gofmt -w .

help:
	@echo "make           :compile the source code"
	@echo "make clean     :remove binary file and vim swp files"
	@echo "make gotool    :run go tool 'fmt' and 'vet'"

.PHONY: clean gotool help
