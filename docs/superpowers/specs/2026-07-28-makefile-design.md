# tabtoy Makefile 设计

- 日期: 2026-07-28
- 状态: 已批准并实现
- 作者: leo

## 背景与目标

tabtoy 当前用 `Make.sh`（bash 脚本）完成跨平台构建与打包。本设计将其转换为 GNU Makefile，在复刻现有行为的基础上做以下改进：

1. **支持交叉架构**（amd64/arm64）构建，修正原脚本 tar 名硬编码 `x86_64` 的不一致。
2. 增加 `clean` 目标清理构建产物。
3. 区分日常构建与发布构建：`make build` 只编当前宿主机平台，`make build_release` 编全部 6 个组合。
4. 增加 `help` 目标，并作为 `make` 默认行为。

`Make.sh` 保留并存，不删除。

## 现状分析

`Make.sh` 行为：

- 硬编码 `Version=3.1.4`。
- 通过 `-ldflags -X` 把 `Version` / `GitCommit` / `BuildTime` 注入 `build` 包（见 `build/build.go` 的 `Version`、`GitCommit`、`BuildTime` 变量）。
- 交叉编译 windows/linux/darwin，只设 `GOOS`，`GOARCH` 跟随宿主机。
- 产物 `bin/<os>/tabtoy<.exe?>`，打包成 `tabtoy-<version>-<os>-x86_64.tar.gz`（`x86_64` 硬编码，与实际架构不符）。
- `set -e` fail-fast；`go build -p 4` 并发编译。
- 无参数构建三平台，带参数构建指定平台。

`bin/` 目录在 `.gitignore` 中（第 32 行），仅 `.gitkeep` 被跟踪；`bin/` 下的 `*.xlsx` 是本地测试数据（未被跟踪，供 `.run` 配置使用），`clean` 必须保留。

`git describe --tags --always` 在当前 HEAD 输出 `3.1.4-35-g224b06a`（HEAD 在 `3.1.4` 标签之后 35 个提交）。

## 设计

### 变量

```makefile
VERSION      := $(shell git describe --tags --always)
BINARY_NAME  := tabtoy
PACKAGE      := github.com/davyxu/tabtoy
BUILD_PKG    := github.com/davyxu/tabtoy/build
BIN_DIR      := bin
BUILD_FLAGS  := -v -p 4
PLATFORMS    := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64 windows/arm64
HOST_OS      := $(shell go env GOOS)
HOST_ARCH    := $(shell go env GOARCH)
BUILD_TIME   := $(shell date -R)
GIT_COMMIT   := $(shell git rev-parse HEAD)
LDFLAGS      := -X "$(BUILD_PKG).BuildTime=$(BUILD_TIME)" \
                -X "$(BUILD_PKG).Version=$(VERSION)" \
                -X "$(BUILD_PKG).GitCommit=$(GIT_COMMIT)"
.DEFAULT_GOAL := help
```

**版本说明**：`git describe --tags --always` 在标签点输出干净版本号（如 `3.1.4`），在标签之后的提交输出带提交定位的串（如 `3.1.4-35-g224b06a`）。标签命名无 `v` 前缀（现有标签为 `3.1.4` 等），故版本串无 `v` 前缀。

**`HOST_OS` / `HOST_ARCH`**：用 `go env GOOS` / `go env GOARCH` 取宿主机平台，供 `make build` 定位当前平台 target。

**`.DEFAULT_GOAL := help`**：`make` 不带参数时显示帮助，而非直接构建。显式指定以避免 `eval` 生成的 `build-darwin-arm64` 等被当成首个目标而成为默认。

### target 结构

用 `define` 模板 + `foreach`/`eval` 生成 6 个 `build-<os>-<arch>` target；`build` 仅构建当前宿主机平台，`build_release` 聚合全部 6 个：

```makefile
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
```

要点：

- **产物路径** `bin/<os>-<arch>/`（如 `bin/darwin-arm64/`），避免 amd64/arm64 在同一 `bin/darwin/` 下撞车。这是相对 `Make.sh`（用 `bin/<os>/`）的必要改动。
- **`.exe` 后缀** 用 `$(if $(filter windows,$(1)),.exe,)` 仅 windows 添加。
- **tar 名** `tabtoy-<version>-<os>-<arch>.tar.gz`，放仓库根（`$(CURDIR)`），如 `tabtoy-3.1.4-35-g224b06a-darwin-arm64.tar.gz`。
- **当前平台构建**：`make build`；**全量构建**：`make build_release` 或 `make -j6 build_release`；**单组合构建**：`make build-linux-arm64`。
- 若宿主机平台不在 `PLATFORMS` 中（如 freebsd），`make build` 会因 `build-<host>` target 不存在而报错；常见宿主（darwin/linux/windows × amd64/arm64）均已覆盖。

### clean

```makefile
.PHONY: clean
clean:
	-@find $(BIN_DIR) -mindepth 1 -maxdepth 1 -type d -exec $(RM) -r {} +
	$(RM) $(BIN_DIR)/$(BINARY_NAME) $(BIN_DIR)/$(BINARY_NAME).exe
	$(RM) $(BINARY_NAME)-*.tar.gz
```

- 第 1 行只删 `bin/` 下的**子目录**（`darwin-arm64/` 等所有构建产物目录），不动 `bin/` 根目录的 `*.xlsx` 与 `.gitkeep`。行首 `-` 忽略该行错误（`bin/` 不存在时 `find` 报错也不中断后续清理）。
- 第 2 行删 `bin/tabtoy`（`.run` 配置 `output_directory` 产生的根二进制）。
- 第 3 行删仓库根的 `tabtoy-*.tar.gz`（含老的 `x86_64` 命名包）。

### help

```makefile
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
```

- 用显式 `@echo` 列出目标，不用 `## `+awk 自动生成方案（`.*?` 非贪婪正则在 BSD awk 下有可移植性风险）。
- `help` 定义在文件末尾；靠 `.DEFAULT_GOAL := help` 使其成为默认，与定义顺序无关。

### 错误处理与并行

- make 天然 fail-fast：任一 recipe 行返回非零即停止，等价 `Make.sh` 的 `set -e`。`go build` 与 `cd && tar` 是两条独立 recipe 行，构建失败时不执行打包。
- 6 个 `build-<os>-<arch>` 互相独立，`make -j6 build_release` 可全并发。每组合内部还跑 `go build -p 4`，6×4 并发编译量较大，机器吃紧时用 `make -j2` 或调小 `BUILD_FLAGS` 的 `-p`。

## 实现注记

- **目标 make 版本**：macOS 自带 GNU Make 3.81。所用语法（`:=`、`define`/`endef`、`$(eval)`、`$(foreach)`、`$(call)`、`$(if)`、`$(filter)`、`$(subst)`、`$(word)`、`$(addprefix)`、`$(CURDIR)`、`$(RM)`、`.DEFAULT_GOAL`）均在 3.81 支持范围内，已 dry-run 验证。
- **`$(subst /,-,$(PLATFORMS))` 而非 `$(PLATFORMS:/=-)`**：替换引用 `$(var:a=b)` 只替换词尾后缀 `a`，`darwin/arm64` 的 `/` 在中间不在词尾，故不生效。必须用 `$(subst /,-,...)` 做全替换。
- **ldflags 用单引号 `-ldflags '$(LDFLAGS)'`**：`date -R` 输出含空格（如 `Tue, 28 Jul 2026 15:43:28 +0800`）。`Make.sh` 靠 bash 变量展开 `"${VersionString}"` 不再重新解析内层引号来保留空格；make 则把变量展开成文本后整行交给 shell，shell 会重新解析内层 `"`，导致按空格拆断。用单引号包裹整段 ldflags，shell 把内层 `"` 当字面量，go 链接器再自行解析内层双引号，与 `Make.sh` 行为一致。
- **纯 Go，无需 CGO**：tabtoy 无 `import "C"`，跨架构编译无需设 `CGO_ENABLED=0`，已实测 `linux/amd64`、`windows/arm64` 等在 macOS 上直接编出。

## 行为对照

| 项 | Make.sh | Makefile |
|---|---|---|
| 版本号 | 硬编码 `3.1.4` | `git describe --tags --always` |
| 架构 | 宿主机 GOARCH | 显式 amd64/arm64 |
| 产物路径 | `bin/<os>/` | `bin/<os>-<arch>/` |
| tar 名 | `tabtoy-<v>-<os>-x86_64.tar.gz` | `tabtoy-<v>-<os>-<arch>.tar.gz` |
| 默认行为 | 构建三平台 | 显示 `help` |
| 当前平台构建 | — | `make build` |
| 全量构建 | 三平台 | `make build_release` (6 组合) |
| 单组合构建 | `Make.sh linux` | `make build-linux-arm64` |
| 清理 | 无 | `make clean` |
| fail-fast | `set -e` | make 内建 |

## 范围外（YAGNI）

- 不加 dev 目标（run/test/fmt/vet/install）。
- 不删除 `Make.sh`。
- 不引入 `--dirty` 标记。
