.PHONY: all build install clean

BINARY_NAME=grub-configurator

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .

# Package and install using makepkg
install:
	makepkg -si

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f *.pkg.tar.zst