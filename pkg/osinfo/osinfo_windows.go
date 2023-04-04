package osinfo

import (
	"dmHC/pkg/cfgparser"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	"github.com/yusufpapurcu/wmi"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var upcard []string

const (
	SecondsPerMinute = 60
	SecondsPerHour   = SecondsPerMinute * 60
	SecondsPerDay    = SecondsPerHour * 24
	df               = "2006-01-02 15:04:05"
)

type OSInfo struct {
	MyCfg cfgparser.Cfile
	MyLog *logrus.Logger
}

func (osinfo *OSInfo) Init(mycfg cfgparser.Cfile, mylog *logrus.Logger) {
	osinfo.MyLog = mylog
	osinfo.MyCfg = mycfg
}

func (osinfo *OSInfo) resolveTime(seconds uint64) (day uint64, hour uint64, minute uint64) {
	var left uint64
	day = seconds / SecondsPerDay
	left = seconds % SecondsPerDay
	hour = left / SecondsPerHour
	left = left % SecondsPerHour
	minute = left / SecondsPerMinute
	return
}

//func in(target string, strarray []string) bool {
//	sort.Strings(strarray)
//	index := sort.SearchStrings(strarray, target)
//	if index < len(strarray) && strarray[index] == target {
//		return true
//	}
//	return false
//}

//func getKeys(m map[string]map[string]interface{}) []string {
//	keys := make([]string, 0, len(m))
//	for k := range m {
//		keys = append(keys, k)
//	}
//	return keys
//}

func (osinfo *OSInfo) GetHostInfo() ([]string, map[string]interface{}) {
	v, err := host.Info()
	if err != nil {
		osinfo.MyLog.Warnf("host info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"主机名", "本地时间", "在线时长", "启动时间", "进程总数", "操作系统类型", "操作系统版本", "内核版本", "内核架构", "虚拟化系统", "虚拟化角色", "主机编号"}

	var mymap map[string]interface{}
	mymap = make(map[string]interface{})
	Hostname := v.Hostname
	NowTime := time.Now().Format(df)
	day, hour, min := osinfo.resolveTime(v.Uptime)
	Uptime := strconv.FormatUint(day, 10) + " day " + strconv.FormatUint(hour, 10) + " hour " + strconv.FormatUint(min, 10) + " min"
	int64ts, _ := strconv.ParseInt(strconv.FormatUint(v.BootTime, 10), 10, 64)
	BootTime := time.Unix(int64ts, 0).Format(df)
	Procs := v.Procs
	OS := v.OS
	Platform := v.Platform
	PlatformFamily := v.PlatformFamily
	//PlatformVersion := v.PlatformVersion
	KernelVersion := v.KernelVersion
	KernelArch := v.KernelArch
	VirtualizationSystem := v.VirtualizationSystem
	VirtualizationRole := v.VirtualizationRole
	HostID := v.HostID

	mymap["主机名"] = Hostname
	mymap["本地时间"] = NowTime
	mymap["在线时长"] = Uptime
	mymap["启动时间"] = BootTime
	mymap["进程总数"] = Procs
	mymap["操作系统类型"] = OS
	mymap["操作系统版本"] = Platform
	mymap["操作系统系列"] = PlatformFamily
	mymap["内核版本"] = KernelVersion
	mymap["内核架构"] = KernelArch
	mymap["虚拟化系统"] = VirtualizationSystem
	mymap["虚拟化角色"] = VirtualizationRole
	mymap["主机编号"] = HostID

	return collst, mymap
}

func (osinfo *OSInfo) GetCpuInfo() ([]string, map[string]interface{}) {
	v, err := cpu.Info()
	if err != nil {
		osinfo.MyLog.Warnf("cpu info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"核数", "制造厂商", "内核版本", "产品型号", "步进编号", "物理编号", "主频速率(MHZ)", "缓冲大小(KB)"}

	//var cpumap map[string]map[string]interface{}
	//cpumap = make(map[string]map[string]interface{})
	var mymap map[string]interface{}
	mymap = make(map[string]interface{})
	//var mylst []map[string]interface{}

	var i int
	//for i = 0; i < len(v); i++ {
	//CPU := v[i].CPU
	VendorID := v[i].VendorID
	Family := v[i].Family
	//Model := v[i].Model
	Stepping := v[i].Stepping
	PhysicalID := v[i].PhysicalID
	//CoreID := v[i].CoreID
	Cores := runtime.NumCPU()
	ModelName := v[i].ModelName
	Mhz := v[i].Mhz
	CacheSize := v[i].CacheSize
	//Flags := v[i].Flags
	//Microcode := v[i].Microcode

	mymap = make(map[string]interface{})
	mymap["制造厂商"] = VendorID
	mymap["内核版本"] = Family
	//mymap["Model"] = Model
	mymap["步进编号"] = Stepping
	mymap["物理编号"] = PhysicalID
	//mymap["核编号"] = CoreID
	mymap["核数"] = Cores
	mymap["产品型号"] = ModelName
	mymap["主频速率(MHZ)"] = Mhz
	mymap["缓冲大小(KB)"] = CacheSize
	//mymap["Flags"] = Flags
	//mymap["微码编号"] = Microcode
	//mymap["CPU编号"] = strconv.Itoa(int(CPU))
	//
	//if in(CoreID, getKeys(cpumap)) {
	//	cpumap[CoreID]["CPU编号"] = reflect.ValueOf(cpumap[CoreID]["CPU编号"]).String() + "," + strconv.Itoa(int(CPU))
	//} else {
	//	cpumap[CoreID] = mymap
	//}
	//}

	//for key := range cpumap {
	//	mylst = append(mylst, cpumap[key])
	//}

	return collst, mymap
}

func (osinfo *OSInfo) GetVMInfo() ([]string, map[string]interface{}) {
	v, err := mem.VirtualMemory()
	if err != nil {
		osinfo.MyLog.Warnf("memory info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"内存总量(MB)", "可用总量(MB)", "已用内存(MB)", "使用率(%)", "空闲内存(MB)", "活动内存(MB)", "非活动内存(MB)",
		"缓冲区内存(MB)", "缓存区内存(MB)", "大页总量", "空闲大页", "保留大页",
		"超载大页", "大页页面大小(MB)"}

	var mymap map[string]interface{}
	mymap = make(map[string]interface{})

	Total := v.Total / 1024 / 1024
	Available := v.Available / 1024 / 1024
	Used := v.Used / 1024 / 1024
	UsedPercent, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", v.UsedPercent), 64)
	Free := v.Free / 1024 / 1024
	Active := v.Active / 1024 / 1024
	Inactive := v.Inactive / 1024 / 1024
	//Wired := v.Wired
	//Laundry := v.Laundry
	Buffers := v.Buffers / 1024 / 1024
	Cached := v.Cached / 1024 / 1024
	//WriteBack := v.WriteBack
	//Dirty := v.Dirty
	//WriteBackTmp := v.WriteBackTmp
	//Shared := v.Shared
	//Slab := v.Slab
	//Sreclaimable := v.Sreclaimable
	//Sunreclaim := v.Sunreclaim
	//SwapCached := v.SwapCached
	//CommitLimit := v.CommitLimit
	//CommittedAS := v.CommittedAS
	//HighTotal := v.HighTotal
	//HighFree := v.HighFree
	//LowTotal := v.LowTotal
	//LowFree := v.LowFree
	//SwapTotal := v.SwapTotal/1024/1024
	//SwapFree := v.SwapFree/1024/1024
	//Mapped := v.Mapped
	//VmallocTotal := v.VmallocTotal
	//VmallocUsed := v.VmallocUsed
	//VmallocChunk := v.VmallocChunk
	HugePagesTotal := v.HugePagesTotal
	HugePagesFree := v.HugePagesFree
	HugePagesRsvd := v.HugePagesRsvd
	HugePagesSurp := v.HugePagesSurp
	HugePageSize := v.HugePageSize / 1024 / 1024

	mymap["内存总量(MB)"] = Total
	mymap["可用总量(MB)"] = Available
	mymap["已用内存(MB)"] = Used
	mymap["使用率(%)"] = UsedPercent
	mymap["空闲内存(MB)"] = Free
	mymap["活动内存(MB)"] = Active
	mymap["非活动内存(MB)"] = Inactive
	//mymap["Wired"] = Wired
	//mymap["Laundry"] = Laundry
	mymap["缓冲区内存(MB)"] = Buffers
	mymap["缓存区内存(MB)"] = Cached
	//mymap["WriteBack"] = WriteBack
	//mymap["Dirty"] = Dirty
	//mymap["WriteBackTmp"] = WriteBackTmp
	//mymap["Shared"] = Shared
	//mymap["Slab"] = Slab
	//mymap["Sreclaimable"] = Sreclaimable
	//mymap["Sunreclaim"] = Sunreclaim
	//mymap["SwapCached"] = SwapCached
	//mymap["CommitLimit"] = CommitLimit
	//mymap["CommittedAS"] = CommittedAS
	//mymap["HighTotal"] = HighTotal
	//mymap["HighFree"] = HighFree
	//mymap["LowTotal"] = LowTotal
	//mymap["LowFree"] = LowFree
	//mymap["交换内存总量(MB)"] = SwapTotal
	//mymap["空闲交换内存(MB)"] = SwapFree
	//mymap["Mapped"] = Mapped
	//mymap["VmallocTotal"] = VmallocTotal
	//mymap["VmallocUsed"] = VmallocUsed
	//mymap["VmallocChunk"] = VmallocChunk
	//mymap["VmallocUsed"] = VmallocUsed
	mymap["大页总量"] = HugePagesTotal
	mymap["空闲大页"] = HugePagesFree
	mymap["保留大页"] = HugePagesRsvd
	mymap["超载大页"] = HugePagesSurp
	mymap["大页页面大小(MB)"] = HugePageSize

	return collst, mymap
}

func (osinfo *OSInfo) GetSwapInfo() ([]string, map[string]interface{}) {
	v, err := mem.SwapMemory()
	if err != nil {
		osinfo.MyLog.Warnf("swap info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v.Total == 0 {
		return nil, nil
	}

	collst := []string{"交换内存总量(MB)", "已用交换内存(MB)", "空闲交换内存(MB)", "使用率(%)", "换入次数", "换出次数",
		"换入页面数", "换出页面数", "缺页次数", "缺页IO次数"}

	var mymap map[string]interface{}
	mymap = make(map[string]interface{})

	Total := v.Total
	Used := v.Used
	Free := v.Free
	UsedPercent, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", v.UsedPercent), 64)

	Sin := v.Sin
	Sout := v.Sout
	PgIn := v.PgIn
	PgOut := v.PgOut
	PgFault := v.PgFault
	PgMajFault := v.PgMajFault

	mymap["交换内存总量(MB)"] = Total / 1024 / 1024
	mymap["已用交换内存(MB)"] = Used / 1024 / 1024
	mymap["空闲交换内存(MB)"] = Free / 1024 / 1024
	mymap["使用率(%)"] = UsedPercent
	mymap["换入次数"] = Sin
	mymap["换出次数"] = Sout
	mymap["换入页面数"] = PgIn
	mymap["换出页面数"] = PgOut
	mymap["缺页次数"] = PgFault
	mymap["缺页IO次数"] = PgMajFault

	return collst, mymap
}

func (osinfo *OSInfo) GetSwapDev() ([]string, []map[string]interface{}) {
	v, err := mem.SwapDevices()
	if err != nil {
		osinfo.MyLog.Warnf("swap devices collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"设备名称", "已用大小(MB)", "空闲大小(MB)"}

	var mymap map[string]interface{}
	var mylst []map[string]interface{}

	var i int
	for i = 0; i < len(v); i++ {
		Name := v[i].Name
		UsedBytes := v[i].UsedBytes / 1024 / 1024
		FreeBytes := v[i].FreeBytes / 1024 / 1024

		mymap = make(map[string]interface{})
		mymap["设备名称"] = Name
		mymap["已用大小(MB)"] = UsedBytes
		mymap["空闲大小(MB)"] = FreeBytes

		mylst = append(mylst, mymap)
	}
	return collst, mylst
}

func (osinfo *OSInfo) GetLoadInfo() ([]string, map[string]interface{}) {
	v, err := load.Avg()
	if err != nil {
		osinfo.MyLog.Warnf("load info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"1秒平均负载", "5秒平均负载", "15秒平均负载"}

	var mymap map[string]interface{}
	mymap = make(map[string]interface{})

	Load1 := v.Load1
	Load5 := v.Load5
	Load15 := v.Load15

	mymap["1秒平均负载"] = Load1
	mymap["5秒平均负载"] = Load5
	mymap["15秒平均负载"] = Load15

	return collst, mymap
}

func (osinfo *OSInfo) GetPartInfo() ([]string, []map[string]interface{}) {
	var err error
	var v []disk.PartitionStat
	v, err = disk.Partitions(true)
	if err != nil {
		osinfo.MyLog.Warnf("partition info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"设备名称", "挂载点", "分区类型", "挂载参数", "分区路径", "分区总空间(MB)", "分区空闲空间(MB)", "分区已用空间(MB)",
		"使用率(%)", "Inodes总量", "已用Inodes", "空闲Inodes", "Inodes使用率(%)"}

	var mymap map[string]interface{}
	var mylst []map[string]interface{}

	var i int
	for i = 0; i < len(v); i++ {
		x, err := disk.Usage(v[i].Mountpoint)
		if err != nil {
			osinfo.MyLog.Warnf("diskusage info collect failed! error:%s\n", err)
			return nil, nil
		}

		if x.Total == 0 {
			continue
		}

		Device := v[i].Device
		Mountpoint := v[i].Mountpoint
		Fstype := v[i].Fstype
		Opts := v[i].Opts
		Path := x.Path
		Total := x.Total / 1024 / 1024
		Free := x.Free / 1024 / 1024
		Used := x.Used / 1024 / 1024
		UsedPercent, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", x.UsedPercent), 64)
		InodesTotal := x.InodesTotal
		InodesUsed := x.InodesUsed
		InodesFree := x.InodesFree
		InodesUsedPercent, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", x.InodesUsedPercent), 64)

		mymap = make(map[string]interface{})
		mymap["设备名称"] = Device
		mymap["挂载点"] = Mountpoint
		mymap["分区类型"] = Fstype
		mymap["挂载参数"] = Opts
		mymap["分区路径"] = Path
		mymap["分区总空间(MB)"] = Total
		mymap["分区空闲空间(MB)"] = Free
		mymap["分区已用空间(MB)"] = Used
		mymap["使用率(%)"] = UsedPercent
		mymap["Inodes总量"] = InodesTotal
		mymap["已用Inodes"] = InodesUsed
		mymap["空闲Inodes"] = InodesFree
		mymap["Inodes使用率(%)"] = InodesUsedPercent

		mylst = append(mylst, mymap)
	}
	return collst, mylst
}

type Win32NetworkAdapter struct {
	AdapterType                 string
	AdapterTypeID               uint16
	Availability                uint16
	Caption                     string
	ConfigManagerErrorCode      uint32
	CreationClassName           string
	Description                 string
	DeviceID                    string
	ErrorDescription            string
	GUID                        string
	Index                       uint32
	InterfaceIndex              uint32
	LastErrorCode               uint32
	MACAddress                  string
	Manufacturer                string
	MaxNumberControlled         uint32
	MaxSpeed                    uint64
	Name                        string
	NetConnectionID             string
	NetConnectionStatus         uint16
	NetworkAddresses            []string
	PermanentAddress            string
	PNPDeviceID                 string
	PowerManagementCapabilities []uint16
	ProductName                 string
	ServiceName                 string
	Speed                       uint64
	Status                      string
	StatusInfo                  uint16
	SystemCreationClassName     string
	SystemName                  string
}

func (osinfo *OSInfo) GetSpeed(ethlst *[]Win32NetworkAdapter, mac string) uint64 {
	for _, value := range *ethlst {
		if value.MACAddress == "" {
			continue
		}

		if strings.Compare(value.MACAddress, strings.ToUpper(mac)) == 0 {
			return value.Speed
		}
	}
	return 0
}

func (osinfo *OSInfo) GetNCInfo() ([]string, []map[string]interface{}) {
	v, err := net.Interfaces()
	if err != nil {
		osinfo.MyLog.Warnf("netcard info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"网卡名称", "MAC地址", "IP地址", "网卡速率(MB)", "最大传输单元", "状态标识"}

	var mymap map[string]interface{}
	var mylst []map[string]interface{}

	var netdetail []Win32NetworkAdapter
	err = wmi.Query("select * from Win32_NetworkAdapter", &netdetail)
	if err != nil {
		osinfo.MyLog.Warnf("netcard detail info collect failed! error:%s", err)
	}
	var i int
	for i = 0; i < len(v); i++ {
		if arrays.Contains(v[i].Flags, "up") > -1 {
			upcard = append(upcard, v[i].Name)
			Speed := osinfo.GetSpeed(&netdetail, v[i].HardwareAddr) / 1000 / 1000
			//Index := v[i].Index
			MTU := v[i].MTU
			Name := v[i].Name
			HardwareAddr := v[i].HardwareAddr
			Flags := v[i].Flags
			Addrs := v[i].Addrs

			mymap = make(map[string]interface{})
			//mymap["Index"] = Index
			mymap["最大传输单元"] = MTU
			mymap["网卡名称"] = Name
			mymap["MAC地址"] = HardwareAddr
			mymap["状态标识"] = Flags
			mymap["IP地址"] = Addrs
			mymap["网卡速率(MB)"] = Speed

			mylst = append(mylst, mymap)
		}
	}
	return collst, mylst
}

func (osinfo *OSInfo) GetNCIO() ([]string, []map[string]interface{}) {
	v, err := net.IOCounters(true)
	if err != nil {
		osinfo.MyLog.Warnf("netcard IO info collect failed! error:%s\n", err)
		return nil, nil
	}
	if v == nil {
		return nil, nil
	}

	collst := []string{"网卡名称", "发送字节数", "接收字节数", "发送包数", "接收包数", "接收包错误数", "发送包错误数",
		"接收丢包数", "发送丢包数", "接收FIFO队列长度", "发送FIFO队列长度"}

	var mymap map[string]interface{}
	var mylst []map[string]interface{}

	var i int
	for i = 0; i < len(v); i++ {
		if arrays.Contains(upcard, v[i].Name) == -1 {
			continue
		}
		Name := v[i].Name
		BytesSent := v[i].BytesSent
		BytesRecv := v[i].BytesRecv
		PacketsSent := v[i].PacketsSent
		PacketsRecv := v[i].PacketsRecv
		Errin := v[i].Errin
		Errout := v[i].Errout
		Dropin := v[i].Dropin
		Dropout := v[i].Dropout
		Fifoin := v[i].Fifoin
		Fifoout := v[i].Fifoout

		mymap = make(map[string]interface{})
		mymap["网卡名称"] = Name
		mymap["发送字节数"] = BytesSent
		mymap["接收字节数"] = BytesRecv
		mymap["发送包数"] = PacketsSent
		mymap["接收包数"] = PacketsRecv
		mymap["接收包错误数"] = Errin
		mymap["发送包错误数"] = Errout
		mymap["接收丢包数"] = Dropin
		mymap["发送丢包数"] = Dropout
		mymap["接收FIFO队列长度"] = Fifoin
		mymap["发送FIFO队列长度"] = Fifoout

		mylst = append(mylst, mymap)
	}
	return collst, mylst
}

//func GetProcInfo(mylog *logrus.Logger) ([]string, []map[string]interface{}) {
//	v, err := process.Processes()
//	if err != nil {
//		mylog.Warnf("process info collect failed! error:%s\n", err)
//		return nil, nil
//	}
//	if v == nil {
//		return nil, nil
//	}
//	var collst []string
//	var mymap map[string]interface{}
//	var mylst []map[string]interface{}
//
//	var i int
//	collst = []string{"Pid", "Parent", "Ctime", "Name", "Cmd", "Env", "CpuPct", "Nice"}
//
//	for i = 0; i < len(v); i++ {
//		skipflag := true
//		if v[i].Pid == 0 {
//			continue
//		}
//		CpuPct, err := v[i].CPUPercent()
//		if err != nil {
//			mylog.Warnf("process cpupct info collect failed! error:%s\n", err)
//			return nil, nil
//		}
//		if CpuPct > 50 {
//			skipflag = false
//		}
//
//		Pid := v[i].Pid
//		Parent, err := v[i].Ppid()
//		if err != nil {
//			mylog.Warnf("process parent info collect failed! error:%s\n", err)
//			return nil, nil
//		}
//		Ctime, err := v[i].CreateTime()
//		if err != nil {
//			mylog.Warnf("process ctime info collect failed! error:%s\n", err)
//			return nil, nil
//		}
//		Name, err := v[i].Name()
//		if err != nil {
//			mylog.Warnf("process name info collect failed! error:%s\n", err)
//			return nil, nil
//		}
//		if strings.Contains(Name, "dmserver") {
//			skipflag = false
//		}
//		Cmd, err := v[i].Cmdline()
//		if err != nil {
//			mylog.Warnf("process cmd info collect failed! error:%s\n", err)
//			return nil, nil
//		}
//		Env, err := v[i].Environ()
//		if err != nil {
//			mylog.Warnf("process env info collect failed! error:%s\n", err)
//			return nil, nil
//		}
//		//Cwd,err := v[i].Cwd()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//CpuAff,err := v[i].CPUAffinity()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		Nice, err := v[i].Nice()
//		if err != nil {
//			mylog.Warnf("process nice info collect failed! error:%s\n", err)
//			return nil, nil
//		}
//		//IONice,err := v[i].IOnice()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//ForeGround,err :=v[i].Foreground()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//BackGround,err :=v[i].Background()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//Uids,err  := v[i].Uids()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//Gids,err  := v[i].Gids()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//
//		//Username,err := v[i].Username()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//
//		//Groups,err  := v[i].Groups()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//NumThreads ,err := v[i].NumThreads()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//NumFDs,err  := v[i].NumFDs()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		//Terminal,err := v[i].Terminal()
//		//if err != nil {
//		//	fmt.Printf("Process Info Collect Error:%s\n", err.Error())
//		//	return nil,nil
//		//}
//		if skipflag {
//			continue
//		}
//		mymap = make(map[string]interface{})
//		mymap["Pid"] = Pid
//		mymap["Parent"] = Parent
//		mymap["Ctime"] = time.UnixMilli(Ctime).Format(df)
//		mymap["Name"] = Name
//		mymap["Cmd"] = Cmd
//		mymap["Env"] = Env
//		//mymap["Cwd"] = Cwd
//		mymap["CpuPct"] = CpuPct
//		//mymap["CpuAff"] = CpuAff
//		mymap["Nice"] = Nice
//		//mymap["IONice"] = IONice
//		//mymap["ForeGround"] = ForeGround
//		//mymap["BackGround"] = BackGround
//		//mymap["Uids"] = Uids
//		//mymap["Gids"] = Gids
//
//		//mymap["Username"] = Username
//		//mymap["Groups"] = Groups
//		//mymap["NumThreads"] = NumThreads
//		//mymap["NumFDs"] = NumFDs
//		//mymap["Terminal"] = Terminal
//
//		//collst = []string {"Pid","Parent","Ctime","Name","Cmd","Env","CpuPct","Nice","Username","Groups","NumThreads","NumFDs","Terminal"}
//
//		mylst = append(mylst, mymap)
//	}
//	return collst, mylst
//}
