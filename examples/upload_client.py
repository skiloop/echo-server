#!/usr/bin/env python3
"""
文件上传客户端示例

使用方法：
python upload_client_example.py <file_path>

示例：
python upload_client_example.py test.txt
"""

import hmac
import hashlib
import os
import time
import sys
import requests

# API配置
API_KEY = os.getenv("AUTH_API_KEY", "your-secret-api-key-here")  # 需要与服务器端的ApiKey一致
SERVER_URL = os.getenv("SERVER_URL", "http://localhost:9012/upload")  # 或 https://localhost:9013/upload


def calculate_signature(api_key, timestamp):
    """计算HMAC-SHA256签名"""
    h = hmac.new(api_key.encode(), timestamp.encode(), hashlib.sha256)
    return h.hexdigest()


def upload_file(file_path, upload_dir=None):
    """上传文件到服务器"""
    # 生成时间戳
    timestamp = str(int(time.time()))
    
    # 计算签名
    signature = calculate_signature(API_KEY, timestamp)
    
    # 设置请求头
    headers = {
        'X-Timestamp': timestamp,
        'X-Signature': signature,
    }
    
    # 准备文件
    try:
        with open(file_path, 'rb') as f:
            files = {'file': f}
            
            # 构建URL（可选：指定上传目录）
            url = SERVER_URL
            if upload_dir:
                url = f"{SERVER_URL}?dir={upload_dir}"
            
            # 发送请求
            response = requests.post(url, files=files, headers=headers)
            
            # 处理响应
            if response.status_code == 200:
                result = response.json()
                print(f"✅ 上传成功!")
                print(f"   文件名: {result.get('filename')}")
                print(f"   大小: {result.get('size')} bytes")
                print(f"   保存路径: {result.get('path')}")
            else:
                print(f"❌ 上传失败!")
                print(f"   状态码: {response.status_code}")
                print(f"   错误信息: {response.text}")
                
    except FileNotFoundError:
        print(f"❌ 文件不存在: {file_path}")
    except Exception as e:
        print(f"❌ 发生错误: {e}")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("使用方法: python upload_client_example.py <file_path> [upload_dir]")
        print("示例: python upload_client_example.py test.txt")
        print("示例: python upload_client_example.py test.txt ./custom_uploads")
        sys.exit(1)
    
    file_path = sys.argv[1]
    upload_dir = sys.argv[2] if len(sys.argv) > 2 else None
    
    upload_file(file_path, upload_dir)

