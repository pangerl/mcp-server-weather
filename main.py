import httpx
import logging
import os
from mcp.server.fastmcp import FastMCP
from pydantic import BaseModel, Field

# --- Configuration ---
# Configure logging
logging.basicConfig(level=logging.INFO)

# Get API configuration from environment variables
QWEATHER_API_HOST = os.getenv("QWEATHER_API_HOST")
QWEATHER_API_KEY = os.getenv("QWEATHER_API_KEY")

# Ensure the API key is set
if not QWEATHER_API_KEY:
    raise ValueError("QWEATHER_API_KEY environment variable is not set. Please get a key from dev.qweather.com.")

# --- MCP Server Setup ---
app = FastMCP(
    title="Weather Server",
    description="A server providing weather information using QWeather API.",
)


# --- Core Request Logic ---
async def request(api_host: str, path: str, params: dict) -> dict:
    """A reusable async request function to interact with the QWeather API."""
    url = f"{api_host.rstrip('/')}{path}"
    headers = {"X-QW-Api-Key": QWEATHER_API_KEY}

    logging.info(f"[INFO] Sending request to {url} with params: {params}")

    async with httpx.AsyncClient() as client:
        try:
            resp = await client.get(url, params=params, headers=headers, timeout=10)
            resp.raise_for_status()
            data = resp.json()
            if data.get("code") != "200":
                raise Exception(f"API Error: code={data.get('code')}, message={resp.text}")
            return data
        except httpx.HTTPStatusError as e:
            logging.error(f"HTTP error occurred: {e.response.status_code} for URL {e.request.url}")
            raise Exception(f"Request failed with status {e.response.status_code}")
        except Exception as e:
            logging.error(f"An unexpected error occurred: {e}")
            raise


# --- API Query Functions ---
async def query_location_id(city: str) -> str:
    """Asynchronously queries the location ID for a given city."""
    params = {"location": city}
    data = await request(QWEATHER_API_HOST, "/geo/v2/city/lookup", params)
    if not data.get("location"):
        raise Exception(f"Could not find location ID for city: {city}. Response: {data}")
    return data["location"][0]["id"]


async def query_qweather_now(location_id: str) -> dict:
    """Asynchronously queries the current weather for a given location ID."""
    params = {"location": location_id}
    data = await request(QWEATHER_API_HOST, "/v7/weather/now", params)
    if not data.get("now"):
        raise Exception(f"Could not find weather data for location ID: {location_id}. Response: {data}")
    return data["now"]


# --- MCP Tool Definition ---
class GetCurrentWeatherInput(BaseModel):
    city: str = Field(..., description="The name of the city to get the weather for, e.g., 'beijing'")


@app.tool()
async def get_current_weather(input: GetCurrentWeatherInput):
    """
    Get the current real-time weather for a specific city.
    It first finds the location ID for the city and then queries the weather.
    """
    try:
        location_id = await query_location_id(input.city)
        logging.info(f"Successfully found location ID for {input.city}: {location_id}")
        weather_data = await query_qweather_now(location_id)
        return weather_data
    except Exception as e:
        logging.error(f"Failed to get weather for {input.city}: {e}")
        return {"error": str(e)}

if __name__ == "__main__":
    # Initialize and run the server
    app.run(transport='stdio')
