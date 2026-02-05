package utils

import (
	"net"
	"sort"
	"testing"
	"time"
)

// TestPingData_GetLossRate 测试丢包率计算功能
// 丢包率 = (发送包数 - 接收包数) / 发送包数
func TestPingData_GetLossRate(t *testing.T) {
	tests := []struct {
		name     string // 测试用例名称
		sended   int    // 发送的包数
		received int    // 接收的包数
		expected float32 // 期望的丢包率
	}{
		{"无丢包", 4, 4, 0},
		{"25%% 丢包", 4, 3, 0.25},
		{"50%% 丢包", 4, 2, 0.5},
		{"75%% 丢包", 4, 1, 0.75},
		{"100%% 丢包", 4, 0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &CloudflareIPData{
				PingData: &PingData{
					Sended:   tt.sended,
					Received: tt.received,
				},
			}
			if got := data.getLossRate(); got != tt.expected {
				t.Errorf("getLossRate() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

// TestPingDelaySet_Sort 测试延迟数据排序功能
// 排序规则：先按丢包率升序，再按延迟升序
func TestPingDelaySet_Sort(t *testing.T) {
	ip1 := net.ParseIP("1.1.1.1")
	ip2 := net.ParseIP("2.2.2.2")
	ip3 := net.ParseIP("3.3.3.3")

	data := PingDelaySet{
		{PingData: &PingData{IP: &net.IPAddr{IP: ip3}, Sended: 4, Received: 2, Delay: 100 * time.Millisecond}}, // 50% 丢包，100ms
		{PingData: &PingData{IP: &net.IPAddr{IP: ip1}, Sended: 4, Received: 4, Delay: 50 * time.Millisecond}},  // 0% 丢包，50ms
		{PingData: &PingData{IP: &net.IPAddr{IP: ip2}, Sended: 4, Received: 4, Delay: 200 * time.Millisecond}}, // 0% 丢包，200ms
	}

	// 按丢包率优先排序，然后按延迟排序
	sort.Sort(data)

	// 1.1.1.1 (0% 丢包，50ms) 应该在第一位
	if data[0].PingData.IP.String() != "1.1.1.1" {
		t.Errorf("expected 1.1.1.1 first, got %s", data[0].PingData.IP.String())
	}
	// 2.2.2.2 (0% 丢包，200ms) 应该在第二位，丢包率相同的情况下延迟更低优先
	if data[1].PingData.IP.String() != "2.2.2.2" {
		t.Errorf("expected 2.2.2.2 second, got %s", data[1].PingData.IP.String())
	}
}

// TestPingDelaySet_FilterDelay 测试延迟范围过滤功能
// 只保留延迟在指定范围内的 IP
func TestPingDelaySet_FilterDelay(t *testing.T) {
	ip1 := net.ParseIP("1.1.1.1")
	ip2 := net.ParseIP("2.2.2.2")
	ip3 := net.ParseIP("3.3.3.3")

	originalMaxDelay := InputMaxDelay
	originalMinDelay := InputMinDelay
	defer func() {
		InputMaxDelay = originalMaxDelay
		InputMinDelay = originalMinDelay
	}()

	// 设置延迟范围 100ms - 200ms
	InputMaxDelay = 200 * time.Millisecond
	InputMinDelay = 100 * time.Millisecond

	data := PingDelaySet{
		{PingData: &PingData{IP: &net.IPAddr{IP: ip1}, Delay: 50 * time.Millisecond}},  // 低于最小值
		{PingData: &PingData{IP: &net.IPAddr{IP: ip2}, Delay: 150 * time.Millisecond}}, // 在范围内
		{PingData: &PingData{IP: &net.IPAddr{IP: ip3}, Delay: 300 * time.Millisecond}}, // 高于最大值
	}

	filtered := data.FilterDelay()

	if len(filtered) != 1 {
		t.Errorf("expected 1 result, got %d", len(filtered))
	}
	if filtered[0].PingData.IP.String() != "2.2.2.2" {
		t.Errorf("expected 2.2.2.2, got %s", filtered[0].PingData.IP.String())
	}
}

// TestPingDelaySet_FilterLossRate 测试丢包率过滤功能
// 只保留丢包率低于指定阈值的 IP
func TestPingDelaySet_FilterLossRate(t *testing.T) {
	ip1 := net.ParseIP("1.1.1.1")
	ip2 := net.ParseIP("2.2.2.2")
	ip3 := net.ParseIP("3.3.3.3")

	originalMaxLossRate := InputMaxLossRate
	defer func() {
		InputMaxLossRate = originalMaxLossRate
	}()

	// 设置最大丢包率为 0.25 (25%)
	InputMaxLossRate = 0.25

	data := PingDelaySet{
		{PingData: &PingData{IP: &net.IPAddr{IP: ip1}, Sended: 4, Received: 4}}, // 0% 丢包
		{PingData: &PingData{IP: &net.IPAddr{IP: ip2}, Sended: 4, Received: 3}}, // 25% 丢包
		{PingData: &PingData{IP: &net.IPAddr{IP: ip3}, Sended: 4, Received: 2}}, // 50% 丢包
	}

	filtered := data.FilterLossRate()

	if len(filtered) != 2 {
		t.Errorf("expected 2 results, got %d", len(filtered))
	}
}

// TestDownloadSpeedSet_Sort 测试下载速度排序功能
// 按下载速度降序排序（速度快的在前）
func TestDownloadSpeedSet_Sort(t *testing.T) {
	ip1 := net.ParseIP("1.1.1.1")
	ip2 := net.ParseIP("2.2.2.2")
	ip3 := net.ParseIP("3.3.3.3")

	data := DownloadSpeedSet{
		{PingData: &PingData{IP: &net.IPAddr{IP: ip3}}, DownloadSpeed: 10},
		{PingData: &PingData{IP: &net.IPAddr{IP: ip1}}, DownloadSpeed: 30},
		{PingData: &PingData{IP: &net.IPAddr{IP: ip2}}, DownloadSpeed: 20},
	}

	// 按下载速度降序排序
	sort.Sort(data)

	// 30MB/s 应该在第一位
	if data[0].PingData.IP.String() != "1.1.1.1" {
		t.Errorf("expected 1.1.1.1 first (highest speed), got %s", data[0].PingData.IP.String())
	}
	// 20MB/s 应该在第二位
	if data[1].PingData.IP.String() != "2.2.2.2" {
		t.Errorf("expected 2.2.2.2 second, got %s", data[1].PingData.IP.String())
	}
	// 10MB/s 应该在第三位
	if data[2].PingData.IP.String() != "3.3.3.3" {
		t.Errorf("expected 3.3.3.3 third, got %s", data[2].PingData.IP.String())
	}
}

// TestPingData_toString 测试数据转换为字符串功能
// 验证 CSV 输出格式的正确性
func TestPingData_toString(t *testing.T) {
	ip := net.ParseIP("1.1.1.1")

	data := &CloudflareIPData{
		PingData: &PingData{
			IP:       &net.IPAddr{IP: ip},
			Sended:   4,
			Received: 3,
			Delay:    150 * time.Millisecond,
			Colo:     "LAX",
		},
		DownloadSpeed: 10 * 1024 * 1024,
	}

	result := data.toString()

	if result[0] != "1.1.1.1" {
		t.Errorf("expected IP 1.1.1.1, got %s", result[0])
	}
	if result[1] != "4" {
		t.Errorf("expected Sended 4, got %s", result[1])
	}
	if result[2] != "3" {
		t.Errorf("expected Received 3, got %s", result[2])
	}
	if result[3] != "0.25" {
		t.Errorf("expected loss rate 0.25, got %s", result[3])
	}
	if result[4] != "150.00" {
		t.Errorf("expected delay 150.00, got %s", result[4])
	}
	if result[5] != "10.00" {
		t.Errorf("expected speed 10.00 MB/s, got %s", result[5])
	}
	if result[6] != "LAX" {
		t.Errorf("expected colo LAX, got %s", result[6])
	}
}

// TestPingData_toString_NAColo 测试空地区码显示为 N/A
// 当 colo 为空时，CSV 输出应该显示 N/A
func TestPingData_toString_NAColo(t *testing.T) {
	ip := net.ParseIP("1.1.1.1")

	data := &CloudflareIPData{
		PingData: &PingData{
			IP:       &net.IPAddr{IP: ip},
			Sended:   4,
			Received: 4,
			Delay:    100 * time.Millisecond,
			Colo:     "", // 空 colo 应该显示为 N/A
		},
		DownloadSpeed: 5 * 1024 * 1024,
	}

	result := data.toString()

	if result[6] != "N/A" {
		t.Errorf("expected N/A for empty colo, got %s", result[6])
	}
}

// TestNoPrintResult 测试是否打印结果的判断逻辑
// 当 PrintNum 为 0 时，不打印结果
func TestNoPrintResult(t *testing.T) {
	tests := []struct {
		name     string
		printNum int
		expected bool
	}{
		{"不打印", 0, true},
		{"打印", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrintNum = tt.printNum
			if got := NoPrintResult(); got != tt.expected {
				t.Errorf("NoPrintResult() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

// TestNoOutput 测试是否输出到文件的判断逻辑
// 当 Output 为空或空格时，不输出文件
func TestNoOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"空字符串", "", true},
		{"空格字符串", " ", true},
		{"文件路径", "result.csv", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Output = tt.output
			if got := noOutput(); got != tt.want {
				t.Errorf("noOutput() = %v, expected %v", got, tt.want)
			}
		})
	}
}
