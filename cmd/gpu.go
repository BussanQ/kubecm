package cmd

import (
	"context"
	"fmt"
	"github.com/BussanQ/kubecm/pkg/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

var xpuType = map[string]v1.ResourceName{
	"NVIDIA":    v1.ResourceName("nvidia.com/gpu"),
	"DCU":       v1.ResourceName("hygon.com/dcu"),
	"ASCEND":    v1.ResourceName("huawei.com/Ascend910"),
	"KUNLUNXIN": v1.ResourceName("kunlunxin.com/xpu"),
}

type GpuCommand struct {
	BaseCommand
}

type gpuCluster struct {
	GpuNum  int       `json:"gpu_num"`
	GpuMem  int       `json:"gpu_mem"`
	GpuUse  int       `json:"gpu_use"`
	GpuPod  string    `json:"gpu_pod"`
	GpuType string    `json:"gpu_type"`
	GpuNode []gpuNode `json:"gpu_node"`
}

type gpuNode struct {
	NodeName string `json:"node_name"`
	Ip       string `json:"ip"`
	GpuType  string `json:"gpu_type"`
	GpuNum   int    `json:"gpu_num"`
	GpuUse   int    `json:"gpu_use"`
	CpuType  string `json:"cpu_type"`
	CpuCores int64  `json:"cpu_cores"`
	Memory   int    `json:"memory"`
}

func (gc *GpuCommand) Init() {
	gc.command = &cobra.Command{
		Use:     "gpu",
		Short:   "print gpu info",
		Long:    "print gpu info",
		Aliases: []string{"g"},
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := getGpu()
			if err != nil {
				return err
			}
			fmt.Printf("%s: %s\n",
				color.RedString("GPU 类型"),
				color.HiWhiteString(g.GpuType))
			fmt.Printf("%s: %s\n",
				color.GreenString("GPU 总量"),
				color.HiWhiteString(strconv.Itoa(g.GpuNum)))
			fmt.Printf("%s: %s\n",
				color.RedString("GPU 使用"),
				color.HiWhiteString(strconv.Itoa(g.GpuUse)))
			fmt.Printf("%s:\n%s",
				color.GreenString("GPU Pod "),
				color.HiWhiteString(g.GpuPod))
			fmt.Printf("%s:\n", color.RedString("GPU Node"))
			for _, node := range g.GpuNode {
				fmt.Printf("%s\n",
					color.HiWhiteString(fmt.Sprintf("%s | %d/%d | %s",
						node.NodeName, node.GpuUse, node.GpuNum, node.Ip)))
			}
			return nil
		},
	}
}

func getGpu() (*gpuCluster, error) {
	clientK8s, err := GetClientSet(cfgFile)
	if err != nil {
		return nil, err
	}
	nodes, errN := clientK8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if errN != nil {
		return nil, errN
	}
	allPod, errP := utils.AllActivePods(clientK8s)
	if errP != nil {
		return nil, errP
	}
	var cluster gpuCluster
	var gpus []gpuNode
	for _, node := range nodes.Items {
		var ok bool
		var gpuNum resource.Quantity
		var gpu gpuNode
		for gpuType, gpuResource := range xpuType {
			gpuNum, ok = node.Status.Capacity[gpuResource]
			if ok {
				cluster.GpuType = gpuType
				gpu.GpuType = gpuType
				break
			}
		}
		if !ok {
			continue
		}
		var ips []string
		for _, address := range node.Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				ips = append(ips, address.Address)
			}
		}
		gpu.Ip = strings.Join(ips, ",")
		gpu.NodeName = node.Name
		gpu.CpuType = node.Labels["kubernetes.io/arch"]
		cpuCap := node.Status.Capacity[v1.ResourceCPU]
		gpu.CpuCores, _ = cpuCap.AsInt64()
		memoryCap := node.Status.Capacity[v1.ResourceMemory]
		memStr := memoryCap.String()
		var memNum = 0
		if strings.Contains(memStr, "Ki") {
			memStr = strings.Replace(memStr, "Ki", "", -1)
			memNum, _ = strconv.Atoi(memStr)
			memNum = memNum / 1024
		}
		gpu.Memory = memNum

		gpuNumI, _ := strconv.Atoi(gpuNum.String())
		for _, pod := range allPod {
			if pod.Spec.NodeName == node.Name {
				gpuUse := utils.GpuInPod(&pod, xpuType[cluster.GpuType])
				if gpuUse > 0 {
					cluster.GpuPod += fmt.Sprintf("%s | %s| %s| %d卡 \n",
						pod.Namespace, pod.Name, pod.Spec.NodeName, gpuUse)
				}
				gpu.GpuUse += int(gpuUse)
			}
		}
		gpus = append(gpus, gpu)
		cluster.GpuNum += gpuNumI
		cluster.GpuUse += gpu.GpuUse
	}
	cluster.GpuNode = gpus
	return &cluster, nil
}
