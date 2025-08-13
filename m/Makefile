# Makefile for Go + Tesseract OCR project

# Tesseract & Leptonica 安装路径 (根据你的 brew 安装路径修改)
PREFIX         := /opt/homebrew
INCLUDE_PATH   := $(PREFIX)/include
LIBRARY_PATH   := $(PREFIX)/lib
TESSDATA_PATH  := $(PREFIX)/share/tessdata
PKG_CONFIG     := $(PREFIX)/lib/pkgconfig

# 项目信息
BINARY_NAME    := app
MAIN_FILE      := main.go

# 环境变量
CGO_ENV := \
	CGO_ENABLED=1 \
	PKG_CONFIG_PATH=$(PKG_CONFIG) \
	TESSDATA_PREFIX=$(TESSDATA_PATH) \
	CPPFLAGS="-I$(INCLUDE_PATH)" \
	LDFLAGS="-L$(LIBRARY_PATH)"

.PHONY: all build run clean

all: build

build:
	@echo "🔨 Building $(BINARY_NAME)..."
	$(CGO_ENV) go build -o $(BINARY_NAME) $(MAIN_FILE)

run:
	@echo "🚀 Running $(BINARY_NAME)..."
	$(CGO_ENV) go run $(MAIN_FILE)

clean:
	@echo "🧹 Cleaning up..."
	rm -f $(BINARY_NAME)
