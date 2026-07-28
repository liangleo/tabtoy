# tabtoy 构建脚本（GNU Make）
# 交叉编译多平台二进制并打包，注入版本信息到 build 包

VERSION      := $(shell git describe --tags --always)
BINARY_NAME  := tabtoy
PACKAGE      := github.com/davyxu/tabtoy
BUILD_PKG    := github.com/davyxu/tabtoy/build
BIN_DIR      := bin
BUILD_FLAGS  := -v -p 4
PLATFORMS    := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

# 宿主机平台（make build 的构建目标）
HOST_OS    := $(shell go env GOOS)
HOST_ARCH  := $(shell go env GOARCH)

BUILD_TIME   := $(shell date -R)
GIT_COMMIT   := $(shell git rev-parse HEAD)
LDFLAGS      := -X "$(BUILD_PKG).BuildTime=$(BUILD_TIME)" \
                -X "$(BUILD_PKG).Version=$(VERSION)" \
                -X "$(BUILD_PKG).GitCommit=$(GIT_COMMIT)"

.DEFAULT_GOAL := help

# 单组合构建模板：$(1)=os  $(2)=arch
define build_platform
.PHONY: build-$(1)-$(2)
build-$(1)-$(2):
	@mkdir -p $(BIN_DIR)/$(1)-$(2)
	GOOS=$(1) GOARCH=$(2) go build $(BUILD_FLAGS) \
		-o $(BIN_DIR)/$(1)-$(2)/$(BINARY_NAME)$(if $(filter windows,$(1)),.exe,) \
		-ldflags '$(LDFLAGS)' $(PACKAGE)
	cd $(BIN_DIR)/$(1)-$(2) && \
		tar zcvf $(CURDIR)/$(BINARY_NAME)-$(VERSION)-$(1)-$(2).tar.gz \
			$(BINARY_NAME)$(if $(filter windows,$(1)),.exe,)
endef

# 把 "darwin/arm64" 拆成 os、arch 两参，生成 6 个 target
$(foreach p,$(PLATFORMS),\
	$(eval $(call build_platform,$(word 1,$(subst /, ,$p)),$(word 2,$(subst /, ,$p)))))

# 构建当前宿主机平台
.PHONY: build
build: build-$(HOST_OS)-$(HOST_ARCH)

# 构建全部 6 个组合（发布用, 支持 make -j6 并行）
.PHONY: build_release
build_release: $(addprefix build-,$(subst /,-,$(PLATFORMS)))

# 清理构建产物，保留 bin/ 下的测试数据(.xlsx)与 .gitkeep
.PHONY: clean
clean:
	-@find $(BIN_DIR) -mindepth 1 -maxdepth 1 -type d -exec $(RM) -r {} +
	$(RM) $(BIN_DIR)/$(BINARY_NAME) $(BIN_DIR)/$(BINARY_NAME).exe
	$(RM) $(BINARY_NAME)-*.tar.gz

# 显示可用目标
.PHONY: help
help:
	@echo "tabtoy Makefile 用法:"
	@echo ""
	@echo "  make                    显示此帮助 (默认)"
	@echo "  make build              构建当前宿主机平台 ($(HOST_OS)/$(HOST_ARCH))"
	@echo "  make build_release      构建全部 6 个 OS×ARCH 组合并打包 (可用 -j6 并行)"
	@echo "  make build-<os>-<arch>  构建单个组合, 如 make build-linux-arm64"
	@echo "                          os: darwin | linux | windows;  arch: amd64 | arm64"
	@echo "  make clean              清理构建产物与 tar, 保留 bin/ 下测试数据(.xlsx)"
	@echo "  make help               显示此帮助"
