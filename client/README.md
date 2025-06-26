# gotty-piko-client

gotty-piko 客户端 - 基于终端的远程协助工具

## 功能特性

- 🚀 基于终端的高效远程协助
- 🌐 支持复杂网络环境
- 🔧 集成 gotty 和 piko 服务
- 📱 支持多平台 (Linux, Windows, macOS)
- 🛡️ 低带宽依赖，避免传统远程桌面的高带宽需求

## 快速开始

### 从预编译版本安装

1. 访问 [GitHub Releases](https://github.com/your-repo/gotty-piko-client/releases)
2. 下载适合您系统的版本
3. 解压并设置执行权限（Linux/macOS）

```bash
# Linux/macOS
chmod +x gottyp-linux-amd64
sudo mv gottyp-linux-amd64 /usr/local/bin/gottyp

# Windows
# 直接运行 .exe 文件即可
```

### 从源码构建

#### 前置要求

- Go 1.21 或更高版本
- Git

#### 构建步骤

```bash
# 克隆仓库
git clone https://github.com/your-repo/gotty-piko-client.git
cd gotty-piko-client

# 下载依赖
go mod download

# 构建当前平台
make build

# 构建所有平台
make build-all
```

## 使用方法

### 基本用法

```bash
# 启动本机模式
gottyp --name=my-server

# 连接到远程服务器
gottyp --name=my-server --remote=192.168.1.100:8088
```

### 命令行参数

| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `--name` | string | ✅ | - | 客户端标识名称 |
| `--remote` | string | ❌ | - | 远程服务器地址 (格式: host:port) |
| `--server-port` | int | ❌ | 8022 | piko服务器端口 |
| `--gotty-port` | int | ❌ | 8080 | gotty本地端口 |

## 开发

### 项目结构

```
client/
├── main.go          # 主程序入口
├── src/
│   ├── config.go    # 配置管理
│   └── service.go   # 服务管理
├── Makefile         # 构建脚本
├── go.mod           # Go模块定义
└── go.sum           # 依赖校验
```

### 开发命令

```bash
# 运行测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 代码格式化
make fmt

# 代码检查
make lint

# 下载依赖
make deps

# 清理构建文件
make clean

# 运行程序
make run

# 查看所有可用命令
make help
```

### 构建多平台版本

```bash
# 构建所有平台
make build-all

# 构建特定平台
make build-linux    # Linux (amd64, arm64)
make build-windows  # Windows (amd64, arm64)
make build-darwin   # macOS (amd64, arm64)
```

## CI/CD

项目使用 GitHub Actions 进行持续集成和部署：

### 工作流

- **CI** (`ci.yml`): 日常代码质量检查
  - 多版本 Go 测试
  - 代码检查 (golangci-lint)
  - 安全扫描 (gosec)
  - 多平台构建

- **Release** (`release.yml`): 自动发布
  - 触发条件：推送标签或手动触发
  - 构建所有平台版本
  - 自动创建 GitHub Release
  - 上传构建产物

### 发布新版本

```bash
# 1. 创建并推送标签
git tag v1.0.0
git push origin v1.0.0

# 2. GitHub Actions 将自动：
#    - 运行测试
#    - 构建所有平台版本
#    - 创建 Release
#    - 上传构建产物
```

### 手动发布

1. 在 GitHub 仓库页面
2. 进入 Actions 标签页
3. 选择 "Build and Release" 工作流
4. 点击 "Run workflow"
5. 输入版本号（如：v1.0.0）
6. 点击 "Run workflow"

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发规范

- 遵循 Go 代码规范
- 添加适当的测试
- 更新文档
- 确保 CI 通过

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

如果您遇到问题或有建议，请：

1. 查看 [Issues](https://github.com/your-repo/gotty-piko-client/issues)
2. 创建新的 Issue
3. 联系维护者

---

**注意**: 请将 `your-repo` 替换为实际的 GitHub 仓库地址。 