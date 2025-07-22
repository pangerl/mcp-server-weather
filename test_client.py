import asyncio
import os
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

# 从环境变量中获取和风天气 API 配置
# 在运行此脚本之前，请确保已设置 QWEATHER_API_KEY 和 QWEATHER_API_HOST
if "QWEATHER_API_KEY" not in os.environ or "QWEATHER_API_HOST" not in os.environ:
    print("错误：请设置 QWEATHER_API_KEY 和 QWEATHER_API_HOST 环境变量。")
    exit(1)

async def main():
    # 传递当前环境变量给子进程
    server_params = StdioServerParameters(
        command="python",
        args=["main.py"],
        env=os.environ.copy()
    )
    async with stdio_client(server_params) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            result = await session.call_tool("get_current_weather", {"input": {"city": "beijing"}})
            print(result)

if __name__ == "__main__":
    asyncio.run(main())
