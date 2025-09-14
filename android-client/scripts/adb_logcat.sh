#!/bin/bash

# ADB Logcat 脚本
# 用于查看 Gottyp Android 客户端的日志

set -e

# 支持指定目标设备
TARGET_DEVICE="${1:-}"

# 如果没有指定设备，使用默认的指定设备
if [ -z "$TARGET_DEVICE" ]; then
    TARGET_DEVICE="adb-91bb2dd8a0274fa4-c8adpd._adb-tls-connect._tcp"
fi

echo "📱 ADB Logcat - Gottyp Android 客户端日志"
echo "🎯 目标设备: $TARGET_DEVICE"

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

echo ""
echo "📋 可用的 ADB Logcat 命令:"
echo "1. 查看 GottypAndroid 标签的日志:"
echo "   adb -s $TARGET_DEVICE logcat -s GottypAndroid"
echo ""
echo "2. 过滤 GottypAndroid 相关日志:"
echo "   adb -s $TARGET_DEVICE logcat | grep GottypAndroid"
echo ""
echo "3. 带时间戳的详细日志:"
echo "   adb -s $TARGET_DEVICE logcat -v time | grep -E '(GottypAndroid|gottyp)'"
echo ""
echo "4. 实时查看所有日志:"
echo "   adb -s $TARGET_DEVICE logcat"
echo ""
echo "5. 清除日志缓冲区:"
echo "   adb -s $TARGET_DEVICE logcat -c"
echo ""

# 直接执行选项3：带时间戳的详细日志
echo "🔍 执行: adb -s $TARGET_DEVICE logcat -v time | grep -E '(GottypAndroid|gottyp)'"
echo "按 Ctrl+C 停止查看日志"
adb -s "$TARGET_DEVICE" logcat -v time | grep -E '(GottypAndroid|gottyp)'
