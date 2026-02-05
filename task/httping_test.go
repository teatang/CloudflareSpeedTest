package task

import (
	"net/http"
	"sync"
	"testing"
)

// TestMapColoMap 测试地区码映射创建功能
// 将 -cfcolo 参数转换为 sync.Map 用于快速查找
func TestMapColoMap(t *testing.T) {
	original := HttpingCFColo
	defer func() { HttpingCFColo = original }()

	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantFirst string
	}{
		{"空字符串", "", 0, ""},
		{"单个地区码", "LAX", 1, "LAX"},
		{"多个地区码", "LAX,SEA,SJC", 3, "LAX"},
		{"混合大小写", "lax,sea", 2, "LAX"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HttpingCFColo = tt.input
			result := MapColoMap()

			if tt.wantLen == 0 {
				if result != nil {
					t.Errorf("MapColoMap() = %v, expected nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("MapColoMap() returned nil")
			}

			if tt.wantLen > 0 {
				var count int
				result.Range(func(key, value any) bool {
					count++
					return true
				})
				if count != tt.wantLen {
					t.Errorf("MapColoMap() count = %d, expected %d", count, tt.wantLen)
				}
			}
		})
	}
}

// TestGetHeaderColo_Cloudflare 测试 Cloudflare CDN 地区码提取
// 从 cf-ray 响应头中提取 IATA 机场三字码
func TestGetHeaderColo_Cloudflare(t *testing.T) {
	header := http.Header{}
	header.Set("server", "cloudflare")
	header.Set("cf-ray", "7bd32409eda7b020-LAX")

	colo := getHeaderColo(header)

	if colo != "LAX" {
		t.Errorf("getHeaderColo() = %s, expected LAX", colo)
	}
}

// TestGetHeaderColo_Cloudflare_EmptyRay 测试 Cloudflare 空 Ray 头
func TestGetHeaderColo_Cloudflare_EmptyRay(t *testing.T) {
	header := http.Header{}
	header.Set("server", "cloudflare")
	// cf-ray 为空

	colo := getHeaderColo(header)

	if colo != "" {
		t.Errorf("getHeaderColo() = %s, expected empty string", colo)
	}
}

// TestGetHeaderColo_CDN77 测试 CDN77 CDN 地区码提取
// 从 x-77-pop 响应头中提取二字国家码
func TestGetHeaderColo_CDN77(t *testing.T) {
	header := http.Header{}
	header.Set("server", "CDN77-Turbo")
	header.Set("x-77-pop", "frankfurtDE")

	colo := getHeaderColo(header)

	if colo != "DE" {
		t.Errorf("getHeaderColo() = %s, expected DE", colo)
	}
}

// TestGetHeaderColo_BunnyCDN 测试 Bunny CDN 地区码提取
// 从 server 响应头中提取二字国家码
func TestGetHeaderColo_BunnyCDN(t *testing.T) {
	header := http.Header{}
	header.Set("server", "BunnyCDN-TW1-1121")

	colo := getHeaderColo(header)

	if colo != "TW" {
		t.Errorf("getHeaderColo() = %s, expected TW", colo)
	}
}

// TestGetHeaderColo_AWSCloudFront 测试 AWS CloudFront 地区码提取
// 从 x-amz-cf-pop 响应头中提取 IATA 机场三字码
func TestGetHeaderColo_AWSCloudFront(t *testing.T) {
	header := http.Header{}
	header.Set("x-amz-cf-pop", "SIN52-P1")

	colo := getHeaderColo(header)

	if colo != "SIN" {
		t.Errorf("getHeaderColo() = %s, expected SIN", colo)
	}
}

// TestGetHeaderColo_Fastly 测试 Fastly CDN 地区码提取
// 从 x-served-by 响应头中提取 IATA 机场三字码（取最后一个）
func TestGetHeaderColo_Fastly(t *testing.T) {
	header := http.Header{}
	header.Set("x-served-by", "cache-fra-etou8220141-FRA")

	colo := getHeaderColo(header)

	if colo != "FRA" {
		t.Errorf("getHeaderColo() = %s, expected FRA", colo)
	}
}

// TestGetHeaderColo_Gcore 测试 Gcore CDN 地区码提取
// 从 x-id-fe 响应头中提取二字城市码（转为大写）
func TestGetHeaderColo_Gcore(t *testing.T) {
	header := http.Header{}
	header.Set("x-id-fe", "fr5-hw-edge-gc17")

	colo := getHeaderColo(header)

	if colo != "FR" {
		t.Errorf("getHeaderColo() = %s, expected FR", colo)
	}
}

// TestGetHeaderColo_Unsupported 测试不支持的 CDN
func TestGetHeaderColo_Unsupported(t *testing.T) {
	header := http.Header{}
	header.Set("server", "nginx")

	colo := getHeaderColo(header)

	if colo != "" {
		t.Errorf("getHeaderColo() = %s, expected empty string for unsupported CDN", colo)
	}
}

// TestGetHeaderColo_Empty 测试空响应头
func TestGetHeaderColo_Empty(t *testing.T) {
	header := http.Header{}

	colo := getHeaderColo(header)

	if colo != "" {
		t.Errorf("getHeaderColo() = %s, expected empty string", colo)
	}
}

// TestFilterColo_NoFilter 测试未设置地区过滤
// 当 HttpingCFColomap 为 nil 时，返回原地区码
func TestFilterColo_NoFilter(t *testing.T) {
	original := HttpingCFColomap
	HttpingCFColomap = nil
	defer func() { HttpingCFColomap = original }()

	colo := "LAX"
	result := (&Ping{}).filterColo(colo)

	if result != colo {
		t.Errorf("filterColo(%s) = %s, expected %s", colo, result, colo)
	}
}

// TestFilterColo_WithFilter 测试地区过滤功能
// 只返回匹配指定地区的 IP
func TestFilterColo_WithFilter(t *testing.T) {
	HttpingCFColomap = &sync.Map{}
	HttpingCFColomap.Store("LAX", "LAX")
	HttpingCFColomap.Store("SEA", "SEA")
	defer func() { HttpingCFColomap = nil }()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"匹配地区码", "LAX", "LAX"},
		{"不匹配地区码", "SJC", ""},
		{"空地区码", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := (&Ping{}).filterColo(tt.input)
			if result != tt.expected {
				t.Errorf("filterColo(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}
