#!/bin/sh

# 定义变量，格式为：${变量名:-默认值}
# 1. 修改：默认 Workers 为 2
WORKERS=${GUNICORN_WORKERS:-2}
# 2. 新增：默认 Threads 为 4
THREADS=${GUNICORN_THREADS:-4}
TIMEOUT=${GUNICORN_TIMEOUT:-300}
BIND=${GUNICORN_BIND:-0.0.0.0:8669}

# 打印一下配置，方便调试 (可选)
echo "Starting Gunicorn with:"
echo "Workers:      $WORKERS"
echo "Threads:      $THREADS"
echo "Worker Class: gthread"
echo "Timeout:      $TIMEOUT"
echo "Bind:         $BIND"

# 使用 exec 启动 gunicorn
# 关键改动：
# 1. 添加了 --threads "$THREADS"
# 2. 添加了 -k gthread (这是必须的，因为默认的 sync worker 不支持线程)
exec gunicorn -w "$WORKERS" \
              --threads "$THREADS" \
              -k gthread \
              -t "$TIMEOUT" \
              -b "$BIND" \
              run:app