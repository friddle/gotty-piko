#!/bin/bash

# 环境检查脚本
# 检查 Go、gomobile、Android SDK 等环境

set -e

echo "🔍 检查环境依赖..."

# 检查是否在正确的目录
if [ ! -d "android" ]; then
    echo "❌ 错误: android/ 目录未找到"
    echo "当前目录: $(pwd)"
    echo "期望目录包含: android/"
    exit 1
fi

# 检查并设置 SDKMAN 环境
if [ -f "$HOME/.sdkman/bin/sdkman-init.sh" ]; then
    echo "📦 初始化 SDKMAN 环境..."
    source "$HOME/.sdkman/bin/sdkman-init.sh"
    
    # 设置 Java 版本
    echo "☕ 使用 SDKMAN 管理的 Java 版本"
    sdk use java 17.0.9-tem
    
    # 设置 Gradle 版本
    echo "🔧 使用 SDKMAN 管理的 Gradle 版本"
    sdk use gradle 8.4
else
    echo "❌ SDKMAN 未安装，请先安装 SDKMAN"
    echo "安装命令: curl -s \"https://get.sdkman.io\" | bash"
    exit 1
fi

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go"
    exit 1
fi

# 检查 gomobile 是否安装
if ! command -v gomobile &> /dev/null; then
    echo "📦 安装 gomobile..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    gomobile init
fi

# 检查 Android SDK 和 NDK
if [ -z "$ANDROID_HOME" ]; then
    echo "❌ ANDROID_HOME 未设置"
    echo "请设置 ANDROID_HOME 指向 Android SDK 路径"
    exit 1
fi

if [ -z "$ANDROID_NDK_HOME" ]; then
    echo "⚠️  ANDROID_NDK_HOME 未设置，使用默认 NDK 路径"
    export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/default"
fi

# 检查 NDK 版本是否兼容
if [ -f "$ANDROID_NDK_HOME/source.properties" ]; then
    NDK_VERSION=$(cat "$ANDROID_NDK_HOME/source.properties" | grep "Pkg.Revision" | cut -d'=' -f2 | tr -d ' ')
    NDK_MAJOR_VERSION=$(echo $NDK_VERSION | cut -d'.' -f1)
    if [ "$NDK_MAJOR_VERSION" -lt 21 ]; then
        echo "⚠️  NDK 版本 $NDK_VERSION 较旧，但继续构建..."
        echo "为获得最佳效果，请从 Android Studio SDK Manager 安装更新的 NDK 版本"
    fi
fi

echo "✅ Go 环境检查通过，版本: $(go version | awk '{print $3}')"
echo "✅ NDK 路径设置: $ANDROID_NDK_HOME"
echo "✅ 环境检查通过"
