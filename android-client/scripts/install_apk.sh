#!/bin/bash

# APK 安装脚本
# 安装并启动 Android 应用

set -e

# 支持指定目标设备
TARGET_DEVICE="${1:-}"

# 如果没有指定设备，尝试使用默认的指定设备
if [ -z "$TARGET_DEVICE" ]; then
    TARGET_DEVICE="adb-91bb2dd8a0274fa4-c8adpd._adb-tls-connect._tcp"
fi

echo "📲 安装 APK..."
echo "🎯 目标设备: $TARGET_DEVICE"

# 检查 APK 文件是否存在
APK_PATH="android/app/build/outputs/apk/debug/app-debug.apk"
if [ ! -f "$APK_PATH" ]; then
    echo "❌ APK 文件未找到: $APK_PATH"
    echo "请先运行构建脚本"
    exit 1
fi

echo "✅ 找到 APK 文件: $APK_PATH"

# 检查 adb 是否可用
if ! command -v adb &> /dev/null; then
    echo "❌ adb 未找到，请确保 Android SDK 工具已安装"
    exit 1
fi

# 检查设备连接
echo "🔍 检查设备连接..."
if [ -n "$TARGET_DEVICE" ]; then
    # 检查指定设备是否连接
    if ! adb devices | grep -q "$TARGET_DEVICE"; then
        echo "❌ 指定设备 $TARGET_DEVICE 未连接"
        echo "可用设备:"
        adb devices
        exit 1
    fi
    echo "✅ 目标设备 $TARGET_DEVICE 已连接"
else
    # 检查是否有任何设备连接
    DEVICES=$(adb devices | grep -v "List of devices" | grep -v "^$" | wc -l)
    if [ "$DEVICES" -eq 0 ]; then
        echo "❌ 未找到连接的 Android 设备"
        echo "请确保设备已连接并启用 USB 调试"
        exit 1
    fi
    echo "✅ 找到 $DEVICES 个连接的设备"
fi

# 安装 APK
echo "📲 安装 APK..."
if [ -n "$TARGET_DEVICE" ]; then
    adb -s "$TARGET_DEVICE" install -r "$APK_PATH"
else
    adb install -r "$APK_PATH"
fi

if [ $? -eq 0 ]; then
    echo "✅ APK 安装成功"
    
    # 启动应用
    echo "🚀 启动 Gottyp Android 客户端..."
    if [ -n "$TARGET_DEVICE" ]; then
        adb -s "$TARGET_DEVICE" shell am start -n com.gottyp.android/.MainActivity
    else
        adb shell am start -n com.gottyp.android/.MainActivity
    fi
    
    echo "🎉 安装完成！"
    echo "Gottyp Android 客户端已安装并启动。"
else
    echo "❌ APK 安装失败"
    exit 1
fi
