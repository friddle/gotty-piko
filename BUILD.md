# 构建说明

本项目已经简化了构建流程，移除了测试、格式化、安全扫描等功能，只保留核心的构建功能。

## Makefile 结构

### 根目录 Makefile
总控Makefile，管理整个项目的构建：

```bash
# 构建所有组件
make build

# 构建客户端
make build-client

# 构建服务端
make build-server

# 构建客户端所有平台
make build-client-all

# 构建服务端多平台
make build-server-multi

# 清理所有构建文件
make clean

# 安装依赖
make deps

# 构建发布版本
make release

# 推送Docker镜像
make push
```

### 客户端 Makefile (client/Makefile)
支持多平台编译的Go客户端：

```bash
# 构建当前平台
make build

# 构建所有平台
make build-all

# 构建特定平台
make build-linux
make build-windows
make build-darwin

# 安装依赖
make deps

# 清理构建文件
make clean
```

### 服务端 Makefile (server/Makefile)
支持Docker构建和部署：

```bash
# 构建Docker镜像
make build

# 构建并推送Docker镜像
make build-push

# 构建多平台Docker镜像
make build-multi

# 构建生产版本
make build-prod

# 运行Docker容器
make run

# 使用docker-compose启动服务
make up
make down
```

## GitHub Actions

项目配置了三个GitHub Actions工作流：

### 1. build.yml
- 触发条件：推送到main/develop分支或PR
- 功能：构建客户端和服务端
- 输出：上传构建产物作为artifacts

### 2. ci.yml
- 触发条件：推送到main/develop分支、PR或发布
- 功能：完整的CI/CD流程
- 包含：构建、集成测试、发布

### 3. release.yml
- 触发条件：发布新版本
- 功能：构建发布版本并上传到GitHub Releases
- 包含：多平台客户端构建、Docker镜像推送

## 使用示例

### 本地开发
```bash
# 构建所有组件
make build

# 只构建客户端
make build-client

# 构建客户端所有平台
make build-client-all

# 启动开发环境
cd server && make up
```

### 发布流程
```bash
# 构建发布版本
make release

# 推送Docker镜像
make push
```

## 注意事项

1. 所有测试、格式化、代码检查功能已移除
2. 安全扫描功能已移除
3. 只保留核心的构建和部署功能
4. GitHub Actions会自动处理CI/CD流程
5. 发布时会自动构建多平台版本并上传到GitHub Releases 