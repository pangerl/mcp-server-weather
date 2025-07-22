import asyncio
import os
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

# 设置和风天气 API 环境变量
os.environ["QWEATHER_API_KEY"] = "xxx"
os.environ["QWEATHER_API_HOST"] = "https://xxx.re.qweatherapi.com"

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
