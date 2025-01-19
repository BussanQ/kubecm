package cmd

import (
	"context"
	"fmt"
	"github.com/BussanQ/kubecm/pkg/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

type GpuCommand struct {
	BaseCommand
}

type gpuCluster struct {
	GpuNum  int    `json:"gpu_num"`
	GpuMem  int    `json:"gpu_mem"`
	GpuUse  int    `json:"gpu_use"`
	GpuPod  string `json:"gpu_pod"`
	GpuType string `json:"gpu_type"`
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
				color.BlueString("GPU 类型"),
				color.HiWhiteString(g.GpuType))
			fmt.Printf("%s: %s\n",
				color.GreenString("GPU 总量"),
				color.HiWhiteString(strconv.Itoa(g.GpuNum)))
			fmt.Printf("%s: %s\n",
				color.RedString("GPU 使用"),
				color.HiWhiteString(strconv.Itoa(g.GpuUse)))
			fmt.Printf("%s:\n%s\n",
				color.GreenString("GPU Pod"),
				color.HiWhiteString(g.GpuPod))
			return nil
		},
	}
}

var xpuType = map[string]v1.ResourceName{
	"NVIDIA": "nvidia.com/gpu",
	"DCU":    "hygon.com/dcu",
	"ASCEND": "huawei.com/Ascend910",
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
	for _, node := range nodes.Items {
		var gpuNum resource.Quantity
		var ok bool
		for k, v := range xpuType {
			gpuNum, ok = node.Status.Capacity[v]
			if ok {
				cluster.GpuType = k
				break
			}
		}
		gpuNumI, _ := strconv.Atoi(gpuNum.String())
		for _, pod := range allPod {
			if pod.Spec.NodeName == node.Name {
				gpuUse := utils.GpuInPod(&pod, xpuType[cluster.GpuType])
				if gpuUse > 0 {
					cluster.GpuPod += pod.Name + "\n"
				}
				cluster.GpuUse += int(gpuUse)
			}
		}
		cluster.GpuNum += gpuNumI
	}
	return &cluster, nil
}
