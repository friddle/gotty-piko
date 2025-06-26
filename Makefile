# Makefile for gotty-piko project
# 总控Makefile，管理整个项目的构建

# 项目信息
PROJECT_NAME=gotty-piko
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 子项目目录
CLIENT_DIR=client
SERVER_DIR=server

# 默认目标
.PHONY: all
all: help

# 构建所有组件
.PHONY: build
build: build-client build-server
	@echo "所有组件构建完成"

# 构建客户端
.PHONY: build-client
build-client:
	@echo "构建客户端..."
	@cd ${CLIENT_DIR} && make build

# 构建服务端
.PHONY: build-server
build-server:
	@echo "构建服务端..."
	@cd ${SERVER_DIR} && make build

# 构建客户端所有平台
.PHONY: build-client-all
build-client-all:
	@echo "构建客户端所有平台..."
	@cd ${CLIENT_DIR} && make build-all

# 构建服务端多平台
.PHONY: build-server-multi
build-server-multi:
	@echo "构建服务端多平台..."
	@cd ${SERVER_DIR} && make build-multi

# 清理所有构建文件
.PHONY: clean
clean: clean-client clean-server
	@echo "所有构建文件清理完成"

# 清理客户端
.PHONY: clean-client
clean-client:
	@echo "清理客户端构建文件..."
	@cd ${CLIENT_DIR} && make clean

# 清理服务端
.PHONY: clean-server
clean-server:
	@echo "清理服务端构建文件..."
	@cd ${SERVER_DIR} && make clean

# 安装依赖
.PHONY: deps
deps: deps-client
	@echo "依赖安装完成"

# 安装客户端依赖
.PHONY: deps-client
deps-client:
	@echo "安装客户端依赖..."
	@cd ${CLIENT_DIR} && make deps

# 构建发布版本
.PHONY: release
release: release-client release-server
	@echo "发布版本构建完成"

# 构建客户端发布版本
.PHONY: release-client
release-client:
	@echo "构建客户端发布版本..."
	@cd ${CLIENT_DIR} && make build-all

# 构建服务端发布版本
.PHONY: release-server
release-server:
	@echo "构建服务端发布版本..."
	@cd ${SERVER_DIR} && make build-prod

# 推送Docker镜像
.PHONY: push
push:
	@echo "推送Docker镜像..."
	@cd ${SERVER_DIR} && make build-push

# 显示项目信息
.PHONY: info
info:
	@echo "项目信息:"
	@echo "  项目名称: ${PROJECT_NAME}"
	@echo "  版本: ${VERSION}"
	@echo "  构建时间: ${BUILD_TIME}"
	@echo "  Git提交: ${GIT_COMMIT}"
	@echo "  客户端目录: ${CLIENT_DIR}"
	@echo "  服务端目录: ${SERVER_DIR}"

# 显示帮助信息
.PHONY: help
help:
	@echo "gotty-piko 项目构建工具"
	@echo ""
	@echo "可用的 Make 目标:"
	@echo ""
	@echo "构建相关:"
	@echo "  build           - 构建所有组件"
	@echo "  build-client    - 构建客户端"
	@echo "  build-server    - 构建服务端"
	@echo "  build-client-all- 构建客户端所有平台"
	@echo "  build-server-multi- 构建服务端多平台"
	@echo ""
	@echo "发布相关:"
	@echo "  release         - 构建发布版本"
	@echo "  push            - 推送Docker镜像"
	@echo ""
	@echo "维护相关:"
	@echo "  clean           - 清理所有构建文件"
	@echo "  deps            - 安装依赖"
	@echo "  info            - 显示项目信息"
	@echo "  help            - 显示此帮助信息" 