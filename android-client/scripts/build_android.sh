#!/bin/bash

# Android 构建脚本
# 构建 Android APK

set -e

echo "📱 构建 Android APK..."

# 构建 Go mobile 库
echo "🔨 构建 Go mobile 库..."
# 由于 NDK 版本兼容性问题跳过 gomobile 构建
echo "⚠️  由于 NDK 版本兼容性问题跳过 gomobile 构建"
echo "在 Android 项目中直接使用现有 Go 代码"

# 构建 Android APK
echo "📱 构建 Android APK..."
cd android

# 使用 SDKMAN 管理的 gradle 构建
echo "使用 SDKMAN 管理的 gradle 构建..."
gradle assembleDebug

# 检查 APK 是否构建成功
APK_PATH="app/build/outputs/apk/debug/app-debug.apk"
if [ ! -f "$APK_PATH" ]; then
    echo "❌ APK 构建失败"
    exit 1
fi

echo "✅ APK 构建成功: $APK_PATH"
echo "APK 文件大小: $(du -h $APK_PATH | cut -f1)"

# 返回根目录
cd ..
