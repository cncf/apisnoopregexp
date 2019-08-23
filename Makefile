GO_LIB_FILES=lib.go
GO_BIN_FILES=cmd/gensql/gensql.go cmd/rmatch/rmatch.go
GO_BIN_CMDS=github.com/ii/apisnoopregexp/cmd/gensql github.com/ii/apisnoopregexp/cmd/rmatch
GO_ENV=CGO_ENABLED=0
GO_BUILD=go build -ldflags '-s -w'
GO_INSTALL=go install -ldflags '-s'
GO_FMT=gofmt -s -w
GO_LINT=golint -set_exit_status
GO_VET=go vet
GO_CONST=goconst
GO_IMPORTS=goimports -w
GO_USEDEXPORTS=usedexports
GO_ERRCHECK=errcheck -asserts -ignore '[FS]?[Pp]rint*'
BINARIES=gensql rmatch
STRIP=strip

all: check ${BINARIES}

gensql: cmd/gensql/gensql.go ${GO_LIB_FILES}
	 ${GO_ENV} ${GO_BUILD} -o gensql cmd/gensql/gensql.go

rmatch: cmd/rmatch/rmatch.go ${GO_LIB_FILES}
	 ${GO_ENV} ${GO_BUILD} -o rmatch cmd/rmatch/rmatch.go

fmt: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_FMT}"

lint: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_LINT}"

vet: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_VET}"

imports: ${GO_BIN_FILES}
	./for_each_go_file.sh "${GO_IMPORTS}"

const: ${GO_BIN_FILES}
	${GO_CONST} ./...

usedexports: ${GO_BIN_FILES}
	${GO_USEDEXPORTS} ./...

errcheck: ${GO_BIN_FILES}
	${GO_ERRCHECK} ./...

check: fmt lint imports vet const usedexports errcheck

install: check ${BINARIES}
	${GO_INSTALL} ${GO_BIN_CMDS}

strip: ${BINARIES}
	${STRIP} ${BINARIES}

clean:
	rm -f ${BINARIES}
