#!/bin/bash

# 定义日志文件名称
BASE_LOG_FILE="kb_sse_"

# 循环10次，从10876到10885的端口范围内启动FastAPI应用
for PORT in {10891..10891}; do
#  # 检查当前端口是否被占用，并获取使用该端口的进程ID
#  PID=$(lsof -i:$PORT -t)
#  # 如果存在使用当前端口的进程，则杀掉这些进程
#  if [ ! -z "$PID" ]; then
#    echo "$PORT端口已被占用，进程ID为$PID，正在尝试杀掉..."
#    kill -9 $PID
#    echo "进程已被杀掉。"
#  fi
  ps -ef | grep $PORT | grep -v grep | awk '{print $2}' | xargs kill -9
  sleep 2

  # 启动FastAPI应用，并将输出重定向到指定的日志文件，同时在后台运行
  echo "正在启动FastAPI应用，端口号为$PORT..."
  echo $BASE_LOG_FILE$PORT.log
    LOG_FILE=$BASE_LOG_FILE$PORT nohup gunicorn know_sse:app -w 5 -k uvicorn.workers.UvicornWorker -b 0.0.0.0:$PORT --timeout 600 --preload > /dev/null 2>&1 &
  echo "FastAPI应用启动成功，日志文件为./logs/$BASE_LOG_FILE$PORT.log。"
done
