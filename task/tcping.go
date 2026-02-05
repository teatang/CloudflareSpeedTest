package task

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/XIU2/CloudflareSpeedTest/utils"
)

const (
	tcpConnectTimeout = time.Second * 1
	maxRoutine        = 1000
	defaultRoutines   = 200
	defaultPort       = 443
	defaultPingTimes  = 4
)

var (
	Routines      = defaultRoutines
	TCPPort   int = defaultPort
	PingTimes int = defaultPingTimes
)

// Ping 结构体：TCP/HTTP ping 测试
type Ping struct {
	wg      *sync.WaitGroup       // 用于等待所有 goroutine 完成
	m       *sync.Mutex           // 互斥锁，保护并发写入
	ips     []*net.IPAddr         // 待测试的 IP 列表
	csv     utils.PingDelaySet    // 测试结果集
	control chan bool             // 控制并发数量的通道
	bar     *utils.Bar            // 进度条
}

// 检查并修正默认参数
func checkPingDefault() {
	if Routines <= 0 {
		Routines = defaultRoutines
	}
	if TCPPort <= 0 || TCPPort >= 65535 {
		TCPPort = defaultPort
	}
	if PingTimes <= 0 {
		PingTimes = defaultPingTimes
	}
}

// 创建新的 Ping 实例
func NewPing() *Ping {
	checkPingDefault()
	ips := loadIPRanges()
	return &Ping{
		wg:      &sync.WaitGroup{},
		m:       &sync.Mutex{},
		ips:     ips,
		csv:     make(utils.PingDelaySet, 0),
		control: make(chan bool, Routines), // 缓冲通道，控制并发数
		bar:     utils.NewBar(len(ips), "可用:", ""),
	}
}

// Run 执行延迟测速
func (p *Ping) Run() utils.PingDelaySet {
	if len(p.ips) == 0 {
		return p.csv
	}
	if Httping {
		utils.Cyan.Printf("开始延迟测速（模式：HTTP, 端口：%d, 范围：%v ~ %v ms, 丢包：%.2f)\n", TCPPort, utils.InputMinDelay.Milliseconds(), utils.InputMaxDelay.Milliseconds(), utils.InputMaxLossRate)
	} else {
		utils.Cyan.Printf("开始延迟测速（模式：TCP, 端口：%d, 范围：%v ~ %v ms, 丢包：%.2f)\n", TCPPort, utils.InputMinDelay.Milliseconds(), utils.InputMaxDelay.Milliseconds(), utils.InputMaxLossRate)
	}
	// 启动多个 goroutine 进行并发测试
	for _, ip := range p.ips {
		p.wg.Add(1)
		p.control <- false // 占用一个并发名额
		go p.start(ip)
	}
	p.wg.Wait()       // 等待所有测试完成
	p.bar.Done()       // 完成进度条
	sort.Sort(p.csv)  // 按丢包率、延迟排序
	return p.csv
}

// start 启动单个 IP 的测试 goroutine
func (p *Ping) start(ip *net.IPAddr) {
	defer p.wg.Done()      // 标记完成
	p.tcpingHandler(ip)    // 执行测试
	<-p.control            // 释放一个并发名额
}

// tcping 执行 TCP 连接测试
// 返回值：连接是否成功、连接耗时
func (p *Ping) tcping(ip *net.IPAddr) (bool, time.Duration) {
	startTime := time.Now()
	var fullAddress string
	if isIPv4(ip.String()) {
		fullAddress = fmt.Sprintf("%s:%d", ip.String(), TCPPort)
	} else {
		fullAddress = fmt.Sprintf("[%s]:%d", ip.String(), TCPPort)
	}
	conn, err := net.DialTimeout("tcp", fullAddress, tcpConnectTimeout)
	if err != nil {
		return false, 0
	}
	defer conn.Close()
	duration := time.Since(startTime)
	return true, duration
}

// checkConnection 检查 IP 连接情况
// 返回值：成功次数、总延迟、地区码
func (p *Ping) checkConnection(ip *net.IPAddr) (recv int, totalDelay time.Duration, colo string) {
	if Httping {
		recv, totalDelay, colo = p.httping(ip)
		return
	}
	colo = "" // TCPing 模式不获取 colo
	// 执行多次 TCP 连接测试
	for i := 0; i < PingTimes; i++ {
		if ok, delay := p.tcping(ip); ok {
			recv++
			totalDelay += delay
		}
	}
	return
}

// appendIPData 线程安全地添加 IP 测试数据
func (p *Ping) appendIPData(data *utils.PingData) {
	p.m.Lock()
	defer p.m.Unlock()
	p.csv = append(p.csv, utils.CloudflareIPData{
		PingData: data,
	})
}

// tcpingHandler 处理单个 IP 的 ping 测试
func (p *Ping) tcpingHandler(ip *net.IPAddr) {
	recv, totalDlay, colo := p.checkConnection(ip)
	nowAble := len(p.csv)
	if recv != 0 {
		nowAble++
	}
	p.bar.Grow(1, strconv.Itoa(nowAble))
	if recv == 0 {
		return
	}
	// 计算平均延迟
	data := &utils.PingData{
		IP:       ip,
		Sended:   PingTimes,
		Received: recv,
		Delay:    totalDlay / time.Duration(recv),
		Colo:     colo,
	}
	p.appendIPData(data)
}
