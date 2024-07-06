package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
	"github.com/sunny0826/kubecm/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GpuCommand struct {
	BaseCommand
}

type gpuCluster struct {
	GpuNum int    `json:"gpu_num"`
	GpuMem int    `json:"gpu_mem"`
	GpuUse int    `json:"gpu_use"`
	GpuPod string `json:"gpu_pod"`
}

func (gc *GpuCommand) Init(){
	gc.command = &cobra.Command{
		Use: "gpu",
		Short: "print gpu info",
		Long: "print gpu info",
		Aliases: []string{"g"},
		RunE: func (cmd *cobra.Command, args []string) error {
			g, err := getGpu()
			if err != nil {return err}
			fmt.Printf("%s: %s\n",
			    ansi.Color("GPU 总量", "green"),
				ansi.Color(strconv.Itoa(g.GpuNum), "white+h"))
			fmt.Printf("%s: %s\n",
			    ansi.Color("GPU 使用", "red"),
				ansi.Color(strconv.Itoa(g.GpuUse), "white+h"))
			fmt.Printf("%s: %s\n",
			    ansi.Color("GPU Pod", "green"),
				ansi.Color(g.GpuPod, "white+h"))
			return nil
		},
	}
}

func getGpu()(*gpuCluster, error){
	clientK8s, err := GetClientSet(cfgFile)
	if err != nil {return nil, err}
	nodes, errN := clientK8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if errN != nil {return nil, errN}
	allPod,errP := utils.AllActivePods(clientK8s)
	if errP != nil {return nil, errP}
	var cluster gpuCluster
	for _, node := range nodes.Items{
		gpuNum, ok := node.Status.Capacity["nvidia.com/gpu"]
		if !ok {
			continue
		}
		gpuNumI, _ := strconv.Atoi(gpuNum.String())
		for _, pod := range allPod {
			if pod.Spec.NodeName == node.Name {
				gpuUse := utils.GpuInPod(&pod)
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


