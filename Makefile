.PHONY: all build install clean test help
all: build
build:
	@bash make.sh build
install:
	@bash make.sh install $(INSTALL_DIR)
clean:
	@bash make.sh clean
test:
	@bash make.sh test
help:
	@echo "make build"
	@echo "make INSTALL_DIR=installDir install"
	@echo "make clean"
	@echo "make test"
