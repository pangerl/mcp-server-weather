package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// 和风天气 API 响应结构体（只保留常用字段）
type QWeatherNowResp struct {
	Code string `json:"code"`
	Now  struct {
		ObsTime   string `json:"obsTime"`
		Temp      string `json:"temp"`
		FeelsLike string `json:"feelsLike"`
		Text      string `json:"text"`
		WindDir   string `json:"windDir"`
		WindScale string `json:"windScale"`
		Humidity  string `json:"humidity"`
		Precip    string `json:"precip"`
		Vis       string `json:"vis"`
		Cloud     string `json:"cloud"`
	} `json:"now"`
	Refer struct {
		Sources []string `json:"sources"`
		License []string `json:"license"`
	} `json:"refer"`
}

type QWeatherCityLookupResp struct {
	Code     string `json:"code"`
	Location []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Adm1    string `json:"adm1"`
		Adm2    string `json:"adm2"`
		Country string `json:"country"`
	} `json:"location"`
}

// 校验 location 是否为 LocationID 或 经纬度
func isLocationIDOrLatLon(location string) bool {
	// LocationID: 全数字，通常为9位
	idRe := regexp.MustCompile(`^\d{6,}$`)
	// 经纬度: 116.41,39.92
	latlonRe := regexp.MustCompile(`^-?\d{1,3}\.\d{1,6},-?\d{1,3}\.\d{1,6}$`)
	return idRe.MatchString(location) || latlonRe.MatchString(location)
}

// 查询城市名称对应的 LocationID
func queryLocationID(apiHost, apiKey, city string, client *http.Client) (string, error) {
	urlStr := fmt.Sprintf("%s/geo/v2/city/lookup?location=%s", strings.TrimRight(apiHost, "/"), url.QueryEscape(city))
	log.Printf("[INFO] 查询城市LocationID: %s", urlStr)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	// 新认证方式
	req.Header.Set("X-QW-Api-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("城市查询请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var lookup QWeatherCityLookupResp
	if err := json.Unmarshal(body, &lookup); err != nil {
		return "", fmt.Errorf("城市查询响应解析失败: %w", err)
	}
	if lookup.Code != "200" || len(lookup.Location) == 0 {
		return "", fmt.Errorf("城市查询失败，code=%s, body=%s", lookup.Code, string(body))
	}
	return lookup.Location[0].ID, nil
}

// 查询实时天气
func queryQWeatherNow(apiHost, apiKey, location string, client *http.Client) (*QWeatherNowResp, error) {
	urlStr := fmt.Sprintf("%s/v7/weather/now?location=%s", strings.TrimRight(apiHost, "/"), url.QueryEscape(location))
	log.Printf("[INFO] 查询天气: %s", urlStr)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	// 新认证方式
	req.Header.Set("X-QW-Api-Key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("天气查询请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var now QWeatherNowResp
	if err := json.Unmarshal(body, &now); err != nil {
		return nil, fmt.Errorf("天气响应解析失败: %w", err)
	}
	if now.Code != "200" {
		return nil, fmt.Errorf("天气查询失败，code=%s, body=%s", now.Code, string(body))
	}
	return &now, nil
}

func getForecast(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("[MCP] get_forecast called, params: %+v", request.Params)
	location, err := request.RequireString("location")
	if err != nil {
		return mcp.NewToolResultError("参数错误: location 必填"), nil
	}
	apiHost := os.Getenv("QWEATHER_API_HOST")
	apiKey := os.Getenv("QWEATHER_API_KEY")
	if apiHost == "" || apiKey == "" {
		return mcp.NewToolResultError("服务未配置和风天气API Host或Key"), nil
	}
	client := &http.Client{Timeout: 8 * time.Second}

	var locationID string
	if isLocationIDOrLatLon(location) {
		locationID = location
	} else {
		locationID, err = queryLocationID(apiHost, apiKey, location, client)
		if err != nil {
			log.Printf("[ERROR] 城市查ID失败: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("城市查ID失败: %v", err)), nil
		}
	}

	weather, err := queryQWeatherNow(apiHost, apiKey, locationID, client)
	if err != nil {
		log.Printf("[ERROR] 天气查询失败: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("天气查询失败: %v", err)), nil
	}
	resp := map[string]any{
		"location":    location,
		"location_id": locationID,
		"weather":     weather,
	}
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[ERROR] 结果序列化失败: %v", err)
		return mcp.NewToolResultError("结果序列化失败"), nil
	}
	log.Printf("[MCP] get_forecast 返回: %s", string(jsonBytes))
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func main() {

	// 创建 mcp-server，名称为 weather
	s := server.NewMCPServer(
		"weather-mcp-server", // server 名称
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	// 注册 get_forecast 工具
	getForecastTool := mcp.NewTool(
		"get_forecast",
		mcp.WithDescription("查询实时天气，location 可为城市名称、LocationID 或经纬度。推荐传城市名，自动查ID。"),
		mcp.WithString("location", mcp.Required(), mcp.Description("查询的地理位置（如城市名称、LocationID或经纬度，推荐城市名)")),
	)

	s.AddTool(getForecastTool, getForecast)

	// 增加统一的 JSON-RPC 日志（仅保留普通日志，回退 ServeStdioWithLogger）
	log.Printf("[MCP] MCP server 启动，等待请求...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
