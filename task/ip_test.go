package task

import (
	"net"
	"testing"
)

// TestIsIPv4 测试 IPv4 地址识别功能
// IPv4 地址包含 "."，IPv6 地址包含 ":"
func TestIsIPv4(t *testing.T) {
	tests := []struct {
		ip     string
		isIPv4 bool
	}{
		{"1.1.1.1", true},
		{"192.168.1.1", true},
		{"255.255.255.255", true},
		{"::1", false},
		{"2606:4700::", false},
		{"2001:db8::1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if got := isIPv4(tt.ip); got != tt.isIPv4 {
				t.Errorf("isIPv4(%s) = %v, expected %v", tt.ip, got, tt.isIPv4)
			}
		})
	}
}

// TestRandIPEndWith 测试随机生成 IP 最后一段功能
// 用于从 IP 段中随机选择一个 IP
func TestRandIPEndWith(t *testing.T) {
	tests := []struct {
		name string
		num  byte
		min  byte
		max  byte
	}{
		{"零值返回零", 0, 0, 0},
		{"最大范围", 255, 0, 254},
		{"小范围", 10, 0, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 运行多次确保在范围内
			for i := 0; i < 100; i++ {
				got := randIPEndWith(tt.num)
				if got < tt.min || got > tt.max {
					t.Errorf("randIPEndWith(%d) = %d, expected between %d and %d", tt.num, got, tt.min, tt.max)
				}
			}
		})
	}
}

// TestIPRanges_FixIP 测试 IP 地址修复功能
// 为没有子网掩码的 IP 添加默认掩码
func TestIPRanges_FixIP(t *testing.T) {
	r := newIPRanges()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"IPv4 无掩码", "1.1.1.1", "1.1.1.1/32"},
		{"IPv6 无掩码", "2606:4700::", "2606:4700::/128"},
		{"IPv4 有掩码", "1.1.1.0/24", "1.1.1.0/24"},
		{"IPv6 有掩码", "2606:4700::/32", "2606:4700::/32"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.fixIP(tt.input)
			if got != tt.expect {
				t.Errorf("fixIP(%s) = %s, expected %s", tt.input, got, tt.expect)
			}
		})
	}
}

// TestIPRanges_GetIPRange 测试获取 IP 范围功能
// 计算子网中可用 IP 的最小值和数量
func TestIPRanges_GetIPRange(t *testing.T) {
	tests := []struct {
		name      string
		ipNet     string
		expectMin byte
		expectMax byte
	}{
		{"/24 子网 192.168.1.0", "192.168.1.0/24", 0, 255},
		{"/25 子网", "10.0.0.0/25", 0, 127},
		{"/30 子网", "10.0.0.0/30", 0, 3},
		{"/28 子网", "10.0.0.0/28", 0, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newIPRanges()
			r.parseCIDR(tt.ipNet)
			minIP, hosts := r.getIPRange()

			if minIP != tt.expectMin {
				t.Errorf("getIPRange() minIP = %d, expected %d", minIP, tt.expectMin)
			}
			if hosts != tt.expectMax {
				t.Errorf("getIPRange() hosts = %d, expected %d", hosts, tt.expectMax)
			}
		})
	}
}

// TestLoadIPRanges_SingleIP 测试加载单个 IP 功能
func TestLoadIPRanges_SingleIP(t *testing.T) {
	// 保存原始值
	originalIPFile := IPFile
	originalIPText := IPText
	defer func() {
		IPFile = originalIPFile
		IPText = originalIPText
	}()

	IPText = "1.1.1.1"
	ips := loadIPRanges()

	if len(ips) != 1 {
		t.Errorf("expected 1 IP, got %d", len(ips))
	}
	if ips[0].String() != "1.1.1.1" {
		t.Errorf("expected 1.1.1.1, got %s", ips[0].String())
	}
}

// TestLoadIPRanges_CIDR 测试加载 CIDR 网段功能
func TestLoadIPRanges_CIDR(t *testing.T) {
	originalIPFile := IPFile
	originalIPText := IPText
	defer func() {
		IPFile = originalIPFile
		IPText = originalIPText
	}()

	// /30 网段包含 4 个 IP
	IPText = "10.0.0.0/30"
	ips := loadIPRanges()

	// 不使用 TestAll 时，只随机获取 1 个 IP
	if len(ips) != 1 {
		t.Errorf("expected 1 random IP, got %d", len(ips))
	}
}

// TestLoadIPRanges_IPv6 测试加载 IPv6 地址功能
func TestLoadIPRanges_IPv6(t *testing.T) {
	originalIPFile := IPFile
	originalIPText := IPText
	defer func() {
		IPFile = originalIPFile
		IPText = originalIPText
	}()

	IPText = "2606:4700::/128"
	ips := loadIPRanges()

	if len(ips) != 1 {
		t.Errorf("expected 1 IPv6 IP, got %d", len(ips))
	}
	if ips[0].String() != "2606:4700::" {
		t.Errorf("expected 2606:4700::, got %s", ips[0].String())
	}
}

// TestIPRanges_AppendIP 测试添加 IP 到列表功能
func TestIPRanges_AppendIP(t *testing.T) {
	r := newIPRanges()
	ip := net.ParseIP("1.1.1.1")
	r.appendIP(ip)

	if len(r.ips) != 1 {
		t.Errorf("expected 1 IP in list, got %d", len(r.ips))
	}
	if r.ips[0].String() != "1.1.1.1" {
		t.Errorf("expected 1.1.1.1, got %s", r.ips[0].String())
	}
}

// TestInitRandSeed 测试初始化随机种子功能
func TestInitRandSeed(t *testing.T) {
	// 验证不 panic
	InitRandSeed()
}
