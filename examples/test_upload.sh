#!/bin/bash

# 文件上传测试脚本

API_KEY="${AUTH_API_KEY:-your-secret-api-key-here}"
SERVER_URL="http://localhost:9012/upload"
TEST_FILE="${TEST_FILE:-test.txt}"

echo "=== 文件上传测试 ==="
echo ""

# 检查测试文件是否存在
if [ ! -f "$TEST_FILE" ]; then
    echo "❌ 测试文件不存在: $TEST_FILE"
    echo "请先创建测试文件"
    exit 1
fi
echo "TEST_FILE : $TEST_FILE"
echo "API_KEY   : $API_KEY"
echo "SERVER_URL: $SERVER_URL"
echo ""

# 生成时间戳
TIMESTAMP=$(date +%s)
echo "时间戳: $TIMESTAMP"

# 计算签名
SIGNATURE=$(printf "%s" "$TIMESTAMP" | openssl dgst -sha256 -hmac "$API_KEY" | awk '{print $2}')
echo "签名: $SIGNATURE"
echo ""

# 上传文件
echo "正在上传文件..."
echo ""

RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$SERVER_URL" \
  -H "X-Timestamp: $TIMESTAMP" \
  -H "X-Signature: $SIGNATURE" \
  -F "file=@$TEST_FILE")

# 分离响应体和状态码
HTTP_BODY=$(echo "$RESPONSE" | sed -e 's/HTTP_STATUS\:.*//g')
HTTP_STATUS=$(echo "$RESPONSE" | tr -d '\n' | sed -e 's/.*HTTP_STATUS://')

echo "HTTP状态码: $HTTP_STATUS"
echo "响应内容: $HTTP_BODY"
echo ""

if [ "$HTTP_STATUS" = "200" ]; then
    echo "✅ 上传成功!"
else
    echo "❌ 上传失败!"
fi

