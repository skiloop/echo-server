#!/bin/bash

echo "=== 子目录上传测试 ==="
echo ""

API_KEY="${AUTH_API_KEY:-your-secret-api-key-here}"
SERVER_URL="${SERVER_URL:-http://localhost:9012/upload}"
TEST_FILE="${TEST_FILE:-test.txt}"

# 检查服务器
if ! curl -s http://localhost:9012/health > /dev/null 2>&1; then
    echo "❌ 服务器未运行，请先启动服务器"
    exit 1
fi

echo "✅ 服务器正在运行"
echo ""

# 检查测试文件
if [ ! -f "$TEST_FILE" ]; then
    echo "创建测试文件..."
    echo "测试内容 - $(date)" > "$TEST_FILE"
fi

# 辅助函数：上传文件到指定子目录
upload_to_dir() {
    local subdir="$1"
    local description="$2"
    
    TIMESTAMP=$(date +%s)
    SIGNATURE=$(printf "%s" "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "测试: $description"
    echo "子目录: ${subdir:-<根目录>}"
    echo ""
    
    local url="$SERVER_URL"
    if [ -n "$subdir" ]; then
        url="${SERVER_URL}?dir=${subdir}"
    fi
    
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$url" \
      -H "X-Timestamp: $TIMESTAMP" \
      -H "X-Signature: $SIGNATURE" \
      -F "file=@$TEST_FILE")
    
    HTTP_BODY=$(echo "$RESPONSE" | sed -e 's/HTTP_STATUS\:.*//g')
    HTTP_STATUS=$(echo "$RESPONSE" | tr -d '\n' | sed -e 's/.*HTTP_STATUS://')
    
    if [ "$HTTP_STATUS" = "200" ]; then
        SAVE_PATH=$(echo "$HTTP_BODY" | grep -o '"path":"[^"]*"' | cut -d'"' -f4)
        FILENAME=$(echo "$HTTP_BODY" | grep -o '"filename":"[^"]*"' | cut -d'"' -f4)
        echo "✅ 上传成功"
        echo "   文件名: $FILENAME"
        echo "   相对路径: $SAVE_PATH"
        echo "   (相对于根上传目录)"
    else
        echo "❌ 上传失败 (状态码: $HTTP_STATUS)"
        echo "   响应: $HTTP_BODY"
    fi
    echo ""
}

# 测试用例

echo "📁 正常子目录测试"
echo ""

# 测试1: 根目录
upload_to_dir "" "上传到根目录"

# 测试2: 单级子目录
upload_to_dir "user1" "上传到 user1 子目录"

# 测试3: 多级子目录
upload_to_dir "user1/photos/2024" "上传到多级子目录"

# 测试4: 项目目录
upload_to_dir "projects/projectA/docs" "上传到项目文档目录"

echo ""
echo "🛡️ 安全测试（路径遍历攻击）"
echo ""

# 测试5: 路径遍历 - 应该被清理
upload_to_dir "../../../etc" "路径遍历攻击 (../../../etc) - 应该被清理为 etc"

# 测试6: 绝对路径 - 应该被清理
upload_to_dir "/etc/config" "绝对路径攻击 (/etc/config) - 应该被清理为 etc/config"

# 测试7: 混合路径遍历
upload_to_dir "user/../../../etc" "混合遍历 (user/../../../etc) - 应该被清理为 etc"

# 测试8: 多个点
upload_to_dir "../../.." "多级遍历 (../../..) - 应该被清理为根目录"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎉 测试完成！"
echo ""
echo "💡 验证要点："
echo "  1. 所有文件都应该保存在根上传目录下"
echo "  2. 子目录路径中的 .. 和 / 应该被清理"
echo "  3. 恶意路径被清理后仍然在根目录内"
echo ""
echo "📂 查看上传的文件："
echo "  ls -R \$(go run . 2>&1 | grep 'Upload root' | awk '{print \$NF}')"
echo ""

