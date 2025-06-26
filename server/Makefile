# Makefile for gotty-piko-server
# 支持Docker构建和部署

# 变量定义
IMAGE_NAME=gotty-piko-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker相关变量
DOCKER_REGISTRY?=ghcr.io
DOCKER_NAMESPACE?=friddle
DOCKER_TAG?=latest

# 构建目录
BUILD_DIR=dist
DIST_DIR=dist

# 默认目标
.PHONY: all
all: clean build

# 构建Docker镜像
.PHONY: build
build:
	@echo "构建 ${IMAGE_NAME} Docker镜像..."
	@mkdir -p ${BUILD_DIR}
	docker build -f build/Dockerfile -t ${IMAGE_NAME}:${DOCKER_TAG} .
	@echo "Docker镜像构建完成: ${IMAGE_NAME}:${DOCKER_TAG}"

# 构建并推送Docker镜像
.PHONY: build-push
build-push: build
	@echo "推送Docker镜像到注册表..."
	@if [ -n "${DOCKER_REGISTRY}" ]; then \
		docker tag ${IMAGE_NAME}:${DOCKER_TAG} ${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${IMAGE_NAME}:${DOCKER_TAG}; \
		docker push ${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${IMAGE_NAME}:${DOCKER_TAG}; \
		echo "镜像已推送到: ${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${IMAGE_NAME}:${DOCKER_TAG}"; \
	else \
		echo "DOCKER_REGISTRY 未设置，跳过推送"; \
	fi

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -rf ${BUILD_DIR} ${DIST_DIR}
	@docker rmi ${IMAGE_NAME}:${DOCKER_TAG} 2>/dev/null || true
	@echo "清理完成"

# 清理所有相关镜像
.PHONY: clean-all
clean-all:
	@echo "清理所有相关Docker镜像..."
	@docker images | grep ${IMAGE_NAME} | awk '{print $$3}' | xargs -r docker rmi -f
	@echo "清理完成"

# 显示镜像信息
.PHONY: info
info:
	@echo "镜像信息:"
	@docker images | grep ${IMAGE_NAME} || echo "未找到相关镜像"

# 进入容器
.PHONY: shell
shell:
	@echo "进入容器shell..."
	@docker exec -it ${IMAGE_NAME}-test /bin/sh

# 构建开发版本
.PHONY: build-dev
build-dev:
	@echo "构建开发版本..."
	docker build -t ${IMAGE_NAME}:dev ${BUILD_DIR}
	@echo "开发版本构建完成: ${IMAGE_NAME}:dev"

# 构建生产版本
.PHONY: build-prod
build-prod:
	@echo "构建生产版本..."
	docker build -t ${IMAGE_NAME}:${VERSION} ${BUILD_DIR}
	docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME}:latest
	@echo "生产版本构建完成: ${IMAGE_NAME}:${VERSION}"

# 帮助
.PHONY: help
help:
	@echo "可用的 Make 目标:"
	@echo "  build         - 构建Docker镜像"
	@echo "  build-push    - 构建并推送Docker镜像"
	@echo "  help          - 显示此帮助信息" 