package task

import (
	"testing"
	"time"
)

// TestCheckPingDefault 测试 TCPing 默认参数校验功能
// 验证参数越界时能正确重置为默认值
func TestCheckPingDefault(t *testing.T) {
	// 保存原始值
	originalRoutines := Routines
	originalTCPPort := TCPPort
	originalPingTimes := PingTimes

	defer func() {
		Routines = originalRoutines
		TCPPort = originalTCPPort
		PingTimes = originalPingTimes
	}()

	tests := []struct {
		name          string
		setup         func()
		wantRoutines  int
		wantTCPPort   int
		wantPingTimes int
	}{
		{
			name: "有效值",
			setup: func() {
				Routines = 500
				TCPPort = 8443
				PingTimes = 6
			},
			wantRoutines:  500,
			wantTCPPort:   8443,
			wantPingTimes: 6,
		},
		{
			name: "零值重置为默认",
			setup: func() {
				Routines = 0
				TCPPort = 443
				PingTimes = 4
			},
			wantRoutines:  defaultRoutines,
			wantTCPPort:   443,
			wantPingTimes: 4,
		},
		{
			name: "负值重置为默认",
			setup: func() {
				Routines = -100
				TCPPort = 443
				PingTimes = 4
			},
			wantRoutines:  defaultRoutines,
			wantTCPPort:   443,
			wantPingTimes: 4,
		},
		{
			name: "无效端口零值重置为默认",
			setup: func() {
				Routines = 200
				TCPPort = 0
				PingTimes = 4
			},
			wantRoutines:  200,
			wantTCPPort:   defaultPort,
			wantPingTimes: 4,
		},
		{
			name: "无效端口最大值重置为默认",
			setup: func() {
				Routines = 200
				TCPPort = 65535
				PingTimes = 4
			},
			wantRoutines:  200,
			wantTCPPort:   defaultPort,
			wantPingTimes: 4,
		},
		{
			name: "零次 ping 重置为默认",
			setup: func() {
				Routines = 200
				TCPPort = 443
				PingTimes = 0
			},
			wantRoutines:  200,
			wantTCPPort:   443,
			wantPingTimes: defaultPingTimes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			checkPingDefault()

			if Routines != tt.wantRoutines {
				t.Errorf("Routines = %d, expected %d", Routines, tt.wantRoutines)
			}
			if TCPPort != tt.wantTCPPort {
				t.Errorf("TCPPort = %d, expected %d", TCPPort, tt.wantTCPPort)
			}
			if PingTimes != tt.wantPingTimes {
				t.Errorf("PingTimes = %d, expected %d", PingTimes, tt.wantPingTimes)
			}
		})
	}
}

// TestNewPing_Defaults 测试新建 Ping 对象功能
func TestNewPing_Defaults(t *testing.T) {
	// 重置为默认值
	Routines = defaultRoutines
	TCPPort = defaultPort
	PingTimes = defaultPingTimes
	IPFile = defaultInputFile
	IPText = ""

	// 临时使用 IPText 避免文件问题
	originalIPFile := IPFile
	originalIPText := IPText
	IPText = "1.1.1.1"
	defer func() {
		IPFile = originalIPFile
		IPText = originalIPText
	}()

	ping := NewPing()
	if ping == nil {
		t.Error("NewPing returned nil")
	}
	if ping.wg == nil {
		t.Error("WaitGroup is nil")
	}
	if ping.m == nil {
		t.Error("Mutex is nil")
	}
	if ping.bar == nil {
		t.Error("Progress bar is nil")
	}
}

// TestConstants 测试常量定义
func TestConstants(t *testing.T) {
	if tcpConnectTimeout != 1*time.Second {
		t.Errorf("tcpConnectTimeout = %v, want 1s", tcpConnectTimeout)
	}
	if maxRoutine != 1000 {
		t.Errorf("maxRoutine = %d, want 1000", maxRoutine)
	}
	if defaultRoutines != 200 {
		t.Errorf("defaultRoutines = %d, want 200", defaultRoutines)
	}
	if defaultPort != 443 {
		t.Errorf("defaultPort = %d, want 443", defaultPort)
	}
	if defaultPingTimes != 4 {
		t.Errorf("defaultPingTimes = %d, want 4", defaultPingTimes)
	}
}
