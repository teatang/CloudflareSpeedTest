package task

import (
	"net"
	"testing"
	"time"
)

// TestCheckDownloadDefault 测试下载参数默认校验功能
// 验证参数越界时能正确重置为默认值
func TestCheckDownloadDefault(t *testing.T) {
	// 保存原始值
	originalURL := URL
	originalTimeout := Timeout
	originalTestCount := TestCount
	originalMinSpeed := MinSpeed

	defer func() {
		URL = originalURL
		Timeout = originalTimeout
		TestCount = originalTestCount
		MinSpeed = originalMinSpeed
	}()

	tests := []struct {
		name      string
		setup     func()
		wantURL   string
		wantTime  time.Duration
		wantCount int
		wantSpeed float64
	}{
		{
			name: "有效值",
			setup: func() {
				URL = "https://test.com"
				Timeout = 5 * time.Second
				TestCount = 5
				MinSpeed = 1.0
			},
			wantURL:   "https://test.com",
			wantTime:  5 * time.Second,
			wantCount: 5,
			wantSpeed: 1.0,
		},
		{
			name: "空 URL 重置为默认",
			setup: func() {
				URL = ""
				Timeout = 10 * time.Second
				TestCount = 10
				MinSpeed = 0.0
			},
			wantURL:    defaultURL,
			wantTime:   10 * time.Second,
			wantCount:  10,
			wantSpeed:  defaultMinSpeed,
		},
		{
			name: "零超时重置为默认",
			setup: func() {
				URL = "https://test.com"
				Timeout = 0
				TestCount = 10
				MinSpeed = 0.0
			},
			wantURL:    "https://test.com",
			wantTime:   defaultTimeout,
			wantCount:  10,
			wantSpeed:  defaultMinSpeed,
		},
		{
			name: "零测试数量重置为默认",
			setup: func() {
				URL = "https://test.com"
				Timeout = 10 * time.Second
				TestCount = 0
				MinSpeed = 0.0
			},
			wantURL:    "https://test.com",
			wantTime:   10 * time.Second,
			wantCount:  defaultTestNum,
			wantSpeed:  defaultMinSpeed,
		},
		{
			name: "负最小速度重置为默认",
			setup: func() {
				URL = "https://test.com"
				Timeout = 10 * time.Second
				TestCount = 10
				MinSpeed = -1.0
			},
			wantURL:    "https://test.com",
			wantTime:   10 * time.Second,
			wantCount:  10,
			wantSpeed:  defaultMinSpeed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			checkDownloadDefault()

			if URL != tt.wantURL {
				t.Errorf("URL = %s, expected %s", URL, tt.wantURL)
			}
			if Timeout != tt.wantTime {
				t.Errorf("Timeout = %v, expected %v", Timeout, tt.wantTime)
			}
			if TestCount != tt.wantCount {
				t.Errorf("TestCount = %d, expected %d", TestCount, tt.wantCount)
			}
			if MinSpeed != tt.wantSpeed {
				t.Errorf("MinSpeed = %f, expected %f", MinSpeed, tt.wantSpeed)
			}
		})
	}
}

// TestGetDialContext_IPv4 测试 IPv4 拨号上下文创建
func TestGetDialContext_IPv4(t *testing.T) {
	ip := net.ParseIP("1.1.1.1")
	TCPPort = 443

	dialCtx := getDialContext(&net.IPAddr{IP: ip})

	if dialCtx == nil {
		t.Fatal("getDialContext returned nil")
	}
}

// TestGetDialContext_IPv6 测试 IPv6 拨号上下文创建
func TestGetDialContext_IPv6(t *testing.T) {
	ip := net.ParseIP("2606:4700::")
	TCPPort = 443

	dialCtx := getDialContext(&net.IPAddr{IP: ip})

	if dialCtx == nil {
		t.Fatal("getDialContext returned nil")
	}
}

// TestPrintDownloadDebugInfo 测试下载调试信息打印功能
// 验证不 panic
func TestPrintDownloadDebugInfo(t *testing.T) {
	ip := net.ParseIP("1.1.1.1")
	err := &net.AddrError{Err: "dial error", Addr: "1.1.1.1:443"}

	// 应该不 panic
	printDownloadDebugInfo(&net.IPAddr{IP: ip}, err, 0, "https://test.com", "", nil)
	printDownloadDebugInfo(&net.IPAddr{IP: ip}, err, 403, "https://test.com", "", nil)
}
