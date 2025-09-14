#!/bin/bash

# Gottyp Android Client Installation Script
# This script orchestrates the build and installation process using modular scripts
# Usage: ./run_install_android.sh [device_id]

set -e

# 获取目标设备ID
TARGET_DEVICE="${1:-}"

# 如果没有指定设备，使用默认的指定设备
if [ -z "$TARGET_DEVICE" ]; then
    TARGET_DEVICE="adb-91bb2dd8a0274fa4-c8adpd._adb-tls-connect._tcp"
fi

echo "🚀 开始 Gottyp Android 客户端安装..."
echo "🎯 目标设备: $TARGET_DEVICE"

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$SCRIPT_DIR/scripts"

# 检查 scripts 目录是否存在
if [ ! -d "$SCRIPTS_DIR" ]; then
    echo "❌ 错误: scripts/ 目录未找到"
    echo "脚本目录: $SCRIPTS_DIR"
    exit 1
fi

# 设置脚本权限
chmod +x "$SCRIPTS_DIR"/*.sh

echo "📋 执行安装步骤..."

# 步骤 1: 环境检查
echo "步骤 1/3: 环境检查"
source "$SCRIPTS_DIR/check_environment.sh"

# 步骤 2: 构建 Android APK
echo "步骤 2/3: 构建 Android APK"
source "$SCRIPTS_DIR/build_android.sh"

# 步骤 3: 安装 APK
echo "步骤 3/3: 安装 APK"
if [ -n "$TARGET_DEVICE" ]; then
    source "$SCRIPTS_DIR/install_apk.sh" "$TARGET_DEVICE"
else
    source "$SCRIPTS_DIR/install_apk.sh"
fi

echo "🎉 所有步骤完成！"
echo "Gottyp Android 客户端已成功安装并启动。"
echo ""
echo "📋 后续操作:"
echo "• 查看应用日志: ./scripts/adb_logcat.sh"
echo "• 指定设备查看日志: ./scripts/adb_logcat.sh [device_id]"
echo "• 常用 adb logcat 命令:"
echo "  - adb -s $TARGET_DEVICE logcat -s GottypAndroid"
echo "  - adb -s $TARGET_DEVICE logcat | grep GottypAndroid"
echo "  - adb -s $TARGET_DEVICE logcat -v time | grep -E '(GottypAndroid|gottyp)'"

