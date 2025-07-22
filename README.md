# Weather Server

本项目是一个基于 FastMCP 框架，集成和风天气（QWeather）API 的天气信息服务。

## 目录结构

```
mcp-server-weather/
├── main.py              # 主服务端代码，定义了 API 逻辑
├── pyproject.toml       # Python 项目依赖和元数据
├── README.md            # 项目说明文档
├── test_client.py       # 测试客户端（如有测试代码）
└── uv.lock              # 依赖锁定文件
```

## 功能简介

- 查询指定城市的实时天气信息
- 自动根据城市名查找 QWeather 的 location ID
- 通过 MCP 工具接口暴露天气查询能力

## 环境依赖

- Python 3.8 及以上
- 依赖包见 `pyproject.toml`，主要包括：
  - httpx
  - pydantic
  - mcp.server.fastmcp

## 获取 QWeather API Key

本服务依赖和风天气（QWeather）的 API Key。请按照以下步骤获取：

1. 访问和风天气开发者平台：[https://dev.qweather.com/docs/api/weather/weather-now](https://dev.qweather.com/docs/api/weather/weather-now)
2. 注册并登录账号。
3. 在控制台创建应用，获取你的 API Key。
4. 将 API Key 设置为环境变量 `QWEATHER_API_KEY`，API Host 设置为 `QWEATHER_API_HOST`（通常为 `https://devapi.qweather.com`）。

示例：

```bash
export QWEATHER_API_HOST="https://devapi.qweather.com"
export QWEATHER_API_KEY="你的和风天气API密钥"
```

详细申请流程和接口参数说明请参考官方文档：[和风天气实时天气API文档](https://dev.qweather.com/docs/api/weather/weather-now)

## 快速开始

1. 安装依赖

```bash
pip install -r requirements.txt
# 或者
pip install .
```

2. 配置环境变量（见下方配置示例，或使用 .env 文件/命令行 -e 传递）

3. 使用 uv 启动服务（推荐）

```bash
uv run main.py
```

4. MCP Client 配置示例

```json
{
  "weather": {
    "disabled": false,
    "timeout": 60,
    "type": "stdio",
    "command": "uv",
    "args": [
      "--directory",
      "/path/to/mcp-server-weather",
      "run",
      "main.py"
    ],
    "env": {
      "QWEATHER_API_KEY": "xxx",
      "QWEATHER_API_HOST": "https://xxx.re.qweatherapi.com"
    }
  }
}
```

如需传统方式启动，也可使用：

```bash
python main.py
```

## API 使用说明

### 工具：get_current_weather

- 输入参数：城市名（如 "beijing"）
- 返回：该城市的实时天气数据（JSON）

#### 示例

请求：

```json
{
  "city": "beijing"
}
```

返回：

```json
{
  "temp": "25",
  "text": "晴",
  ...
}
```

返回字段说明详见[官方文档](https://dev.qweather.com/docs/api/weather/weather-now)。

## 错误处理

- 若未设置 `QWEATHER_API_KEY`，程序会直接报错退出。
- 查询失败时，返回格式为 `{"error": "错误描述"}`。

## 参考

- [和风天气实时天气API文档](https://dev.qweather.com/docs/api/weather/weather-now)
