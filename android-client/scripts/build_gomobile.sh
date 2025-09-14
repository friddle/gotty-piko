#!/bin/bash

# Gottyp Gomobile 绑定生成脚本

set -e

echo "🚀 开始生成 Gottyp Gomobile 绑定..."

# 检查gomobile是否安装
if ! command -v gomobile &> /dev/null; then
    echo "❌ gomobile 未安装，正在安装..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    gomobile init
fi

# 检查Android SDK是否配置
if [ -z "$ANDROID_HOME" ]; then
    echo "❌ ANDROID_HOME 环境变量未设置"
    echo "请设置 ANDROID_HOME 指向 Android SDK 目录"
    exit 1
fi

# 检查NDK是否安装
if [ ! -d "$ANDROID_HOME/ndk" ]; then
    echo "❌ Android NDK 未安装"
    echo "请通过 Android Studio 安装 NDK"
    exit 1
fi

echo "✅ 环境检查通过"

# 创建输出目录
mkdir -p dist
mkdir -p android/app/libs

# 清理旧的绑定文件
echo "🧹 清理旧的绑定文件..."
rm -f android/app/libs/gottyp.aar
rm -f dist/gottyp.aar

# 生成gomobile绑定
echo "📦 生成gomobile绑定..."
# 使用gottyp包进行绑定
gomobile bind -target=android -o android/app/libs/gottyp.aar -ldflags="-s -w" ./gottyp

if [ $? -eq 0 ]; then
    echo "✅ Gomobile绑定生成成功"
    echo "📦 AAR文件位置: android/app/libs/gottyp.aar"
    
    # 显示AAR文件信息
    if [ -f "android/app/libs/gottyp.aar" ]; then
        echo "📊 AAR文件大小: $(du -h android/app/libs/gottyp.aar | cut -f1)"
    fi
    
    echo ""
    echo "🎯 下一步操作："
    echo "1. 在Android Studio中同步项目"
    echo "2. 取消注释MainActivity.java中的gomobile调用代码"
    echo "3. 构建并运行Android应用"
    
else
    echo "❌ Gomobile绑定生成失败"
    exit 1
fi

echo "🎉 Gomobile绑定生成完成！"
