#!/bin/bash

# gottyp Android 部署脚本
# 功能：推送gottyp-android到设备，创建并启动服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
REMOTE_SERVER="https://remote-codie-test.code27.cn"
ANDROID_BINARY_DIR="/data/local/tmp"
SERVICE_NAME="gottyp"
SERVICE_DIR="/system/bin"
DEVICE_ID=""  # 将在check_adb中设置

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 执行adb命令的辅助函数
adb_cmd() {
    if [ -n "$DEVICE_ID" ]; then
        adb -s "$DEVICE_ID" "$@"
    else
        adb "$@"
    fi
}

# 检查adb是否可用
check_adb() {
    if ! command -v adb &> /dev/null; then
        log_error "adb命令未找到，请确保Android SDK已安装并添加到PATH"
        exit 1
    fi
    
    # 检查设备连接
    local device_count=$(adb devices | grep "device$" | wc -l)
    if [ "$device_count" -eq 0 ]; then
        log_error "未找到已连接的Android设备，请确保设备已连接并启用USB调试"
        exit 1
    elif [ "$device_count" -gt 1 ]; then
        log_warning "检测到多个设备连接，将使用第一个设备"
        DEVICE_ID=$(adb devices | grep "device$" | head -1 | awk '{print $1}')
        log_info "使用设备: $DEVICE_ID"
    else
        DEVICE_ID=$(adb devices | grep "device$" | awk '{print $1}')
        log_info "使用设备: $DEVICE_ID"
    fi
    
    log_success "adb连接正常"
}

# 获取SNI（从设备主机名或设备ID获取）
get_sni() {
    local device_id=$(adb_cmd shell getprop ro.product.model 2>/dev/null | tr -d '\r\n' || echo "android-device")
    local serial=$(adb_cmd shell getprop ro.serialno 2>/dev/null | tr -d '\r\n' || echo "unknown")
    
    # 组合设备信息作为SNI
    local sni="${device_id}-${serial}"
    # 清理特殊字符，只保留字母数字和连字符
    sni=$(echo "$sni" | sed 's/[^a-zA-Z0-9-]/-/g' | sed 's/--*/-/g' | sed 's/^-\|-$//g')
    
    if [ -z "$sni" ] || [ "$sni" = "-" ]; then
        sni="android-$(date +%s)"
    fi
    
    echo "$sni"
}

# 选择适合的Android二进制文件
select_android_binary() {
    local api_level=$(adb_cmd shell getprop ro.build.version.sdk 2>/dev/null | tr -d '\r\n' || echo "28")
    local arch=$(adb_cmd shell getprop ro.product.cpu.abi 2>/dev/null | tr -d '\r\n' || echo "arm64-v8a")
    
    # 根据API Level选择二进制文件
    local binary_name="gottyp-android-arm64-api${api_level}"
    local binary_path="./dist/${binary_name}"
    
    if [ ! -f "$binary_path" ]; then
        # 尝试使用可用的二进制文件
        for api in 35 34 33 32 31 30 29 28; do
            binary_name="gottyp-android-arm64-api${api}"
            binary_path="./dist/${binary_name}"
            if [ -f "$binary_path" ]; then
                break
            fi
        done
        
        if [ ! -f "$binary_path" ]; then
            return 1
        fi
    fi
    
    echo "$binary_path"
}

# 推送二进制文件到设备
push_binary() {
    local binary_path="$1"
    local remote_path="${ANDROID_BINARY_DIR}/gottyp"
    
    log_info "推送二进制文件到设备..."
    log_info "本地文件: $binary_path"
    log_info "远程路径: $remote_path"
    
    # 推送文件
    if ! adb_cmd push "$binary_path" "$remote_path"; then
        log_error "推送二进制文件失败"
        exit 1
    fi
    
    # 设置执行权限
    if ! adb_cmd shell "suks root chmod 755 $remote_path" 2>/dev/null; then
        log_warning "suks命令不可用，尝试直接设置权限"
        if ! adb_cmd shell "chmod 755 $remote_path"; then
            log_warning "无法设置执行权限，但文件已推送成功"
        else
            log_success "执行权限设置成功"
        fi
    else
        log_success "执行权限设置成功"
    fi
    
    log_success "二进制文件推送成功"
}

# 创建systemd服务文件（如果支持）
create_service() {
    local sni="$1"
    local service_file="/etc/systemd/system/${SERVICE_NAME}.service"
    local binary_path="${ANDROID_BINARY_DIR}/gottyp"
    
    log_info "创建systemd服务..."
    
    # 创建服务文件内容
    local service_content="[Unit]
Description=Gottyp Remote Terminal Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=${binary_path} --name=${sni} --remote=${REMOTE_SERVER}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target"

    # Android设备通常不支持systemd
    log_warning "Android设备不支持systemd，将使用直接启动方式"
    return 1
}

# 启动服务
start_service() {
    local sni="$1"
    local binary_path="${ANDROID_BINARY_DIR}/gottyp"
    
    log_info "启动gottyp服务..."
    
    # 尝试停止现有服务
    log_info "停止现有服务..."
    adb_cmd shell "suks root pkill -f gottyp" 2>/dev/null || adb_cmd shell "pkill -f gottyp" 2>/dev/null || true
    
    # Android设备使用suks启动方式
    log_info "使用suks启动服务..."
    
    # 在后台启动服务
    if ! adb_cmd shell "suks root sh -c 'nohup ${binary_path} --name=${sni} --remote=${REMOTE_SERVER} --insecure-skip-verify > /data/local/tmp/gottyp.log 2>&1 &'" 2>/dev/null; then
        log_warning "suks启动失败，尝试直接启动"
        adb_cmd shell "nohup ${binary_path} --name=${sni} --remote=${REMOTE_SERVER} --insecure-skip-verify > /data/local/tmp/gottyp.log 2>&1 &"
    fi
    
    # 等待一下确保服务启动
    sleep 3
    
    # 检查服务是否运行
    if adb_cmd shell "suks root pgrep -f gottyp" > /dev/null 2>&1 || adb_cmd shell "pgrep -f gottyp" > /dev/null 2>&1; then
        log_success "服务启动成功"
    else
        log_error "服务启动失败"
        log_info "查看日志："
        adb_cmd shell "cat /data/local/tmp/gottyp.log" 2>/dev/null || true
        exit 1
    fi
}

# 显示服务状态
show_status() {
    local sni="$1"
    
    log_info "服务状态信息："
    echo "  服务名称: $SERVICE_NAME"
    echo "  客户端标识: $sni"
    echo "  远程服务器: $REMOTE_SERVER"
    echo "  二进制路径: ${ANDROID_BINARY_DIR}/gottyp"
    
    # 检查进程
    log_info "检查运行状态..."
    if adb_cmd shell "suks root pgrep -f gottyp" > /dev/null 2>&1; then
        log_success "gottyp进程正在运行（使用suks）"
        adb_cmd shell "suks root ps | grep gottyp" 2>/dev/null || true
    elif adb_cmd shell "pgrep -f gottyp" > /dev/null 2>&1; then
        log_success "gottyp进程正在运行"
        adb_cmd shell "ps | grep gottyp" 2>/dev/null || true
    else
        log_warning "未找到gottyp进程"
    fi
    
    # 显示日志
    log_info "最近日志："
    adb_cmd shell "tail -20 /data/local/tmp/gottyp.log" 2>/dev/null || log_warning "无法读取日志文件"
}

# 主函数
main() {
    log_info "开始部署gottyp到Android设备..."
    
    # 检查adb
    check_adb
    
    # 获取SNI
    local sni=$(get_sni)
    log_info "使用SNI: $sni"
    
    # 选择二进制文件
    local binary_path=$(select_android_binary)
    if [ $? -ne 0 ]; then
        log_error "未找到任何可用的Android二进制文件"
        exit 1
    fi
    
    # 获取设备信息用于日志显示
    local api_level=$(adb_cmd shell getprop ro.build.version.sdk 2>/dev/null | tr -d '\r\n' || echo "28")
    local arch=$(adb_cmd shell getprop ro.product.cpu.abi 2>/dev/null | tr -d '\r\n' || echo "arm64-v8a")
    log_info "检测到设备信息：API Level: $api_level, 架构: $arch"
    log_info "使用二进制文件: $binary_path"
    
    # 推送二进制文件
    push_binary "$binary_path"
    
    # 创建服务
    create_service "$sni" || true
    
    # 启动服务（无论服务创建是否成功都尝试启动）
    start_service "$sni"
    
    # 显示状态
    show_status "$sni"
    
    log_success "gottyp部署完成！"
    log_info "服务已启动，客户端标识为: $sni"
    log_info "远程服务器: $REMOTE_SERVER"
}

# 脚本入口
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
