package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	vgpuInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nvidia_vgpu_info",
		Help: "Information about vGPU instances",
	}, []string{"vgpu_id", "vm_name", "vgpu_profile", "driver_version"})

	vgpuMemory = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nvidia_vgpu_memory_bytes",
		Help: "vGPU memory metrics",
	}, []string{"vgpu_id", "vm_name", "type"})

	vgpuUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nvidia_vgpu_utilization_percent",
		Help: "vGPU utilization metrics",
	}, []string{"vgpu_id", "vm_name", "type"})
)

func init() {
	prometheus.MustRegister(vgpuInfo)
	prometheus.MustRegister(vgpuMemory)
	prometheus.MustRegister(vgpuUtilization)
}

func main() {
	user := os.Getenv("APP_USER")
	host := os.Getenv("APP_HOST")
	password := os.Getenv("APP_PASSWORD")

	if user == "" || host == "" || password == "" {
		log.Fatal("Missing required environment variables")
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":9834", nil))
	}()

	for {
		output := runSSHCommand(user, host, password, "nvidia-smi vgpu -q")
		parseVGPUInfo(output)
		time.Sleep(15 * time.Second)
	}
}

func runSSHCommand(user, host, password, command string) string {
	cmdStr := fmt.Sprintf("sshpass -p '%s' ssh -o StrictHostKeyChecking=no %s@%s '%s'",
		password, user, host, command)

	out, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		log.Printf("SSH error: %v", err)
		return ""
	}
	return string(out)
}

func parseVGPUInfo(output string) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var (
		currentVGPU    string
		currentVMName  string
		currentProfile string
		driverVersion  string
		inMemory       bool
		inUtilization  bool
	)

	vgpuInfo.Reset()
	vgpuMemory.Reset()
	vgpuUtilization.Reset()

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "vGPU ID"):
			currentVGPU = strings.Split(trimmed, ":")[1]
			currentVGPU = strings.TrimSpace(currentVGPU)
			inMemory = false
			inUtilization = false

		case strings.HasPrefix(trimmed, "VM Name"):
			currentVMName = strings.Split(trimmed, ":")[1]
			currentVMName = strings.TrimSpace(currentVMName)

		case strings.HasPrefix(trimmed, "vGPU Name"):
			currentProfile = strings.Split(trimmed, ":")[1]
			currentProfile = strings.TrimSpace(currentProfile)

		case strings.HasPrefix(trimmed, "Guest Driver Version"):
			driverVersion = strings.Split(trimmed, ":")[1]
			driverVersion = strings.TrimSpace(driverVersion)
			vgpuInfo.WithLabelValues(currentVGPU, currentVMName, currentProfile, driverVersion).Set(1)

		case strings.HasPrefix(trimmed, "FB Memory Usage"):
			inMemory = true
			inUtilization = false

		case strings.HasPrefix(trimmed, "Utilization"):
			inUtilization = true
			inMemory = false

		default:
			if inMemory && currentVGPU != "" {
				parseMemory(trimmed, currentVGPU, currentVMName)
			}
			if inUtilization && currentVGPU != "" {
				parseUtilization(trimmed, currentVGPU, currentVMName)
			}
		}
	}
}

func parseMemory(line, vgpuID, vmName string) {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	value = strings.TrimSuffix(value, " MiB")

	if mem, err := strconv.ParseFloat(value, 64); err == nil {
		memBytes := mem * 1024 * 1024
		vgpuMemory.WithLabelValues(vgpuID, vmName, key).Set(memBytes)
	}
}

func parseUtilization(line, vgpuID, vmName string) {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	value = strings.TrimSuffix(value, " %")

	if util, err := strconv.ParseFloat(value, 64); err == nil {
		vgpuUtilization.WithLabelValues(vgpuID, vmName, key).Set(util)
	}
}
