#!/bin/bash
# start.sh

export QWEATHER_API_KEY="xxx"
export QWEATHER_API_HOST="https://xxx.re.qweatherapi.com"

# 获取脚本所在的目录，确保我们总是在正确的目录下运行
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$DIR"

# 运行 Go 程序 (可以是 go run，也可以是预编译的二进制文件)
# 方式A: 使用 go run
go run main.go

# 方式B: 运行预编译文件 (更推荐)
# ./weather_service 

