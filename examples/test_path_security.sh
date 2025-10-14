#!/bin/bash

echo "=== 文件上传路径安全测试 ==="
echo ""

API_KEY="${AUTH_API_KEY:-your-secret-api-key-here}"
SERVER_URL="${SERVER_URL:-http://localhost:9012/upload}"

# 检查服务器是否运行
if ! curl -s http://localhost:9012/health > /dev/null 2>&1; then
    echo "❌ 服务器未运行，请先启动服务器"
    exit 1
fi

echo "✅ 服务器正在运行"
echo ""

# 创建测试文件
echo "测试内容" > /tmp/test_upload_security.txt

# 辅助函数：上传文件
upload_file() {
    local filename="$1"
    local description="$2"
    
    TIMESTAMP=$(date +%s)
    SIGNATURE=$(printf "%s" "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "测试: $description"
    echo "文件名: $filename"
    echo ""
    
    RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$SERVER_URL" \
      -H "X-Timestamp: $TIMESTAMP" \
      -H "X-Signature: $SIGNATURE" \
      -F "file=@/tmp/test_upload_security.txt;filename=$filename")
    
    HTTP_BODY=$(echo "$RESPONSE" | sed -e 's/HTTP_STATUS\:.*//g')
    HTTP_STATUS=$(echo "$RESPONSE" | tr -d '\n' | sed -e 's/.*HTTP_STATUS://')
    
    echo "状态码: $HTTP_STATUS"
    
    if [ "$HTTP_STATUS" = "200" ]; then
        SAVE_PATH=$(echo "$HTTP_BODY" | grep -o '"path":"[^"]*"' | cut -d'"' -f4)
        SAVE_FILENAME=$(echo "$HTTP_BODY" | grep -o '"filename":"[^"]*"' | cut -d'"' -f4)
        echo "✅ 上传成功"
        echo "   保存文件名: $SAVE_FILENAME"
        echo "   保存路径: $SAVE_PATH"
        
        # 检查路径是否安全
        if echo "$SAVE_PATH" | grep -q "uploads"; then
            echo "   ✅ 路径安全：在 uploads 目录内"
        else
            echo "   ⚠️  路径可能不安全：$SAVE_PATH"
        fi
    elif [ "$HTTP_STATUS" = "400" ]; then
        ERROR=$(echo "$HTTP_BODY" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
        echo "❌ 请求被拒绝（符合预期）"
        echo "   错误: $ERROR"
    else
        echo "❌ 其他错误"
        echo "   响应: $HTTP_BODY"
    fi
    
    echo ""
}

# 测试1: 正常文件名
upload_file "normal_test.txt" "正常文件名"

# 测试2: 路径遍历攻击 - ../
upload_file "../../../etc/passwd" "路径遍历攻击 (../../../etc/passwd)"

# 测试3: 路径遍历攻击 - 混合
upload_file "test/../../../etc/passwd" "路径遍历攻击 (test/../../../etc/passwd)"

# 测试4: 路径遍历攻击 - 多级
upload_file "....//....//etc/passwd" "路径遍历攻击 (....//....//etc/passwd)"

# 测试5: 绝对路径攻击
upload_file "/etc/cron.d/malicious" "绝对路径攻击 (/etc/cron.d/malicious)"

# 测试6: 当前目录
upload_file "." "特殊文件名 (.)"

# 测试7: 父目录
upload_file ".." "特殊文件名 (..)"

# 测试8: 空文件名
upload_file "" "空文件名"

# 测试9: 包含子目录的文件名
upload_file "subdir/test.txt" "包含子目录 (subdir/test.txt)"

# 测试10: Windows风格路径
upload_file "..\\..\\..\\windows\\system32\\config\\sam" "Windows路径遍历"

# 清理
rm -f /tmp/test_upload_security.txt

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎉 安全测试完成！"
echo ""
echo "💡 预期结果："
echo "  - 正常文件名: 上传成功，保存在 uploads/ 目录"
echo "  - 路径遍历攻击: 文件名被清理，仍然保存在 uploads/ 目录"
echo "  - 特殊文件名 (., ..): 被拒绝"
echo "  - 空文件名: 被拒绝"
echo ""
echo "🔒 所有测试用例都应该保证文件只能保存在 uploads 目录内！"
echo ""

