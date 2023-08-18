package nodes

import (
	"encoding/json"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/events"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/lists"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"os"
	"runtime"
	"strings"
	"time"
)

type NodeStatusExecutor struct {
	isFirstTime bool

	cpuUpdatedTime   time.Time
	cpuLogicalCount  int
	cpuPhysicalCount int
}

func NewNodeStatusExecutor() *NodeStatusExecutor {
	return &NodeStatusExecutor{}
}

func (this *NodeStatusExecutor) Listen() {
	this.isFirstTime = true
	this.cpuUpdatedTime = time.Now()
	this.update()

	// TODO 这个时间间隔可以配置
	ticker := time.NewTicker(30 * time.Second)

	events.On(events.EventQuit, func() {
		remotelogs.Println("NODE_STATUS", "quit executor")
		ticker.Stop()
	})

	for range ticker.C {
		this.isFirstTime = false
		this.update()
	}
}

func (this *NodeStatusExecutor) update() {
	if sharedAPIConfig == nil {
		return
	}

	status := &nodeconfigs.NodeStatus{}
	status.BuildVersion = teaconst.Version
	status.BuildVersionCode = utils.VersionToLong(teaconst.Version)
	status.OS = runtime.GOOS
	status.Arch = runtime.GOARCH
	exe, _ := os.Executable()
	status.ExePath = exe
	status.ConfigVersion = 0
	status.IsActive = true
	status.ConnectionCount = 0 // TODO 实现连接数计算

	hostname, _ := os.Hostname()
	status.Hostname = hostname

	this.updateCPU(status)
	this.updateMem(status)
	this.updateLoad(status)
	this.updateDisk(status)
	status.UpdatedAt = time.Now().Unix()

	//  发送数据
	jsonData, err := json.Marshal(status)
	if err != nil {
		remotelogs.Error("NODE_STATUS", "serial NodeStatus fail: "+err.Error())
		return
	}
	err = models.SharedAPINodeDAO.UpdateAPINodeStatus(nil, sharedAPIConfig.NumberId(), jsonData)
	if err != nil {
		remotelogs.Error("NODE_STATUS", "rpc UpdateNodeStatus() failed: "+err.Error())
		return
	}
}

// 更新CPU
func (this *NodeStatusExecutor) updateCPU(status *nodeconfigs.NodeStatus) {
	duration := time.Duration(0)
	if this.isFirstTime {
		duration = 100 * time.Millisecond
	}
	percents, err := cpu.Percent(duration, false)
	if err != nil {
		status.Error = "cpu.Percent(): " + err.Error()
		return
	}
	if len(percents) == 0 {
		return
	}
	status.CPUUsage = percents[0] / 100

	if time.Since(this.cpuUpdatedTime) > 300*time.Second { // 每隔5分钟才会更新一次
		this.cpuUpdatedTime = time.Now()

		status.CPULogicalCount, err = cpu.Counts(true)
		if err != nil {
			status.Error = "cpu.Counts(): " + err.Error()
			return
		}
		status.CPUPhysicalCount, err = cpu.Counts(false)
		if err != nil {
			status.Error = "cpu.Counts(): " + err.Error()
			return
		}
		this.cpuLogicalCount = status.CPULogicalCount
		this.cpuPhysicalCount = status.CPUPhysicalCount
	} else {
		status.CPULogicalCount = this.cpuLogicalCount
		status.CPUPhysicalCount = this.cpuPhysicalCount
	}
}

// 更新硬盘
func (this *NodeStatusExecutor) updateDisk(status *nodeconfigs.NodeStatus) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		remotelogs.Error("NODE_STATUS", err.Error())
		return
	}
	lists.Sort(partitions, func(i int, j int) bool {
		p1 := partitions[i]
		p2 := partitions[j]
		return p1.Mountpoint > p2.Mountpoint
	})

	// 当前TeaWeb所在的fs
	var rootFS = ""
	var rootTotal = uint64(0)
	if lists.ContainsString([]string{"darwin", "linux", "freebsd"}, runtime.GOOS) {
		for _, p := range partitions {
			if p.Mountpoint == "/" {
				rootFS = p.Fstype
				usage, _ := disk.Usage(p.Mountpoint)
				if usage != nil {
					rootTotal = usage.Total
				}
				break
			}
		}
	}

	var total = rootTotal
	var totalUsage = uint64(0)
	var maxUsage = float64(0)
	for _, partition := range partitions {
		if runtime.GOOS != "windows" && !strings.Contains(partition.Device, "/") && !strings.Contains(partition.Device, "\\") {
			continue
		}

		// 跳过不同fs的
		if len(rootFS) > 0 && rootFS != partition.Fstype {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		if partition.Mountpoint != "/" && (usage.Total != rootTotal || total == 0) {
			total += usage.Total
		}
		totalUsage += usage.Used
		if usage.UsedPercent >= maxUsage {
			maxUsage = usage.UsedPercent
			status.DiskMaxUsagePartition = partition.Mountpoint
		}
	}
	status.DiskTotal = total
	if total > 0 {
		status.DiskUsage = float64(totalUsage) / float64(total)
	}
	status.DiskMaxUsage = maxUsage / 100
}
