/*
 * Copyright (c) 2022, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */
package scraper

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	statsapi "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

type Scraper struct {
	tokenPath string
	timeout   time.Duration
	targetIP  string
	errors    *prometheus.Desc
	errCnt    float64
	logger    *zap.Logger

	nodeFsUsedBytes                            *prometheus.Desc
	nodeFsAvailableBytes                       *prometheus.Desc
	nodeFsInodesFree                           *prometheus.Desc
	nodeFsInodes                               *prometheus.Desc
	nodeFsInodesUsed                           *prometheus.Desc
	nodeRuntimeImageFsUsedBytes                *prometheus.Desc
	nodeRuntimeImageFsAvailableBytes           *prometheus.Desc
	nodeRuntimeImageFsInodesFree               *prometheus.Desc
	nodeRuntimeImageFsInodes                   *prometheus.Desc
	nodeRuntimeImageFsInodesUsed               *prometheus.Desc
	nodeRuntimeContainerFsUsedBytes            *prometheus.Desc
	nodeRuntimeContainerFsAvailableBytes       *prometheus.Desc
	nodeRuntimeContainerFsInodesFree           *prometheus.Desc
	nodeRuntimeContainerFsInodes               *prometheus.Desc
	nodeRuntimeContainerFsInodesUsed           *prometheus.Desc
	nodeCPUUsageNanoCores                      *prometheus.Desc
	nodeCPUUsageCoreNanoSeconds                *prometheus.Desc
	nodeMemoryAvailableBytes                   *prometheus.Desc
	nodeMemoryUsageBytes                       *prometheus.Desc
	nodeMemoryWorkingSetBytes                  *prometheus.Desc
	nodeMemoryRSSBytes                         *prometheus.Desc
	nodeMemoryPageFaults                       *prometheus.Desc
	nodeMemoryMajorPageFaults                  *prometheus.Desc
	nodeSwapAvailableBytes                     *prometheus.Desc
	nodeSwapUsageBytes                         *prometheus.Desc
	nodeRLimitMaxPID                           *prometheus.Desc
	nodeRLimitNumOfRunningProcess              *prometheus.Desc
	nodeInterfaceRxBytes                       *prometheus.Desc
	nodeInterfaceRxErrors                      *prometheus.Desc
	nodeInterfaceTxBytes                       *prometheus.Desc
	nodeInterfaceTxErrors                      *prometheus.Desc
	nodeSystemContainerRootFsUsedBytes         *prometheus.Desc
	nodeSystemContainerRootFsAvailableBytes    *prometheus.Desc
	nodeSystemContainerRootFsInodesFree        *prometheus.Desc
	nodeSystemContainerRootFsInodes            *prometheus.Desc
	nodeSystemContainerRootFsInodesUsed        *prometheus.Desc
	nodeSystemContainerLogsUsedBytes           *prometheus.Desc
	nodeSystemContainerLogsAvailableBytes      *prometheus.Desc
	nodeSystemContainerLogsInodesFree          *prometheus.Desc
	nodeSystemContainerLogsInodes              *prometheus.Desc
	nodeSystemContainerLogsInodesUsed          *prometheus.Desc
	nodeSystemContainerCPUUsageNanoCores       *prometheus.Desc
	nodeSystemContainerCPUUsageCoreNanoSeconds *prometheus.Desc
	nodeSystemContainerMemoryAvailableBytes    *prometheus.Desc
	nodeSystemContainerMemoryUsageBytes        *prometheus.Desc
	nodeSystemContainerMemoryWorkingSetBytes   *prometheus.Desc
	nodeSystemContainerMemoryRSSBytes          *prometheus.Desc
	nodeSystemContainerMemoryPageFaults        *prometheus.Desc
	nodeSystemContainerMemoryMajorPageFaults   *prometheus.Desc
	nodeSystemContainerSwapAvailableBytes      *prometheus.Desc
	nodeSystemContainerSwapUsageBytes          *prometheus.Desc
	nodeSystemContainerAcceleratorMemoryTotal  *prometheus.Desc
	nodeSystemContainerAcceleratorMemoryUsed   *prometheus.Desc
	nodeSystemContainerAcceleratorDutyCycle    *prometheus.Desc

	podCPUUsageNanoCores              *prometheus.Desc
	podCPUUsageCoreNanoSeconds        *prometheus.Desc
	podMemoryAvailableBytes           *prometheus.Desc
	podMemoryUsageBytes               *prometheus.Desc
	podMemoryWorkingSetBytes          *prometheus.Desc
	podMemoryRSSBytes                 *prometheus.Desc
	podMemoryPageFaults               *prometheus.Desc
	podMemoryMajorPageFaults          *prometheus.Desc
	podSwapAvailableBytes             *prometheus.Desc
	podSwapUsageBytes                 *prometheus.Desc
	podInterfaceRxBytes               *prometheus.Desc
	podInterfaceRxErrors              *prometheus.Desc
	podInterfaceTxBytes               *prometheus.Desc
	podInterfaceTxErrors              *prometheus.Desc
	podEphemeralStorageUsedBytes      *prometheus.Desc
	podEphemeralStorageAvailableBytes *prometheus.Desc
	podEphemeralStorageInodesFree     *prometheus.Desc
	podEphemeralStorageInodes         *prometheus.Desc
	podEphemeralStorageInodesUsed     *prometheus.Desc
	podVolumeUsedBytes                *prometheus.Desc
	podVolumeAvailableBytes           *prometheus.Desc
	podVolumeInodesFree               *prometheus.Desc
	podVolumeInodes                   *prometheus.Desc
	podVolumeInodesUsed               *prometheus.Desc
	podVolumeHealthStatus             *prometheus.Desc
	podProcessCount                   *prometheus.Desc

	containerRootFsUsedBytes         *prometheus.Desc
	containerRootFsAvailableBytes    *prometheus.Desc
	containerRootFsInodesFree        *prometheus.Desc
	containerRootFsInodes            *prometheus.Desc
	containerRootFsInodesUsed        *prometheus.Desc
	containerLogsUsedBytes           *prometheus.Desc
	containerLogsAvailableBytes      *prometheus.Desc
	containerLogsInodesFree          *prometheus.Desc
	containerLogsInodes              *prometheus.Desc
	containerLogsInodesUsed          *prometheus.Desc
	containerCPUUsageNanoCores       *prometheus.Desc
	containerCPUUsageCoreNanoSeconds *prometheus.Desc
	containerMemoryAvailableBytes    *prometheus.Desc
	containerMemoryUsageBytes        *prometheus.Desc
	containerMemoryWorkingSetBytes   *prometheus.Desc
	containerMemoryRSSBytes          *prometheus.Desc
	containerMemoryPageFaults        *prometheus.Desc
	containerMemoryMajorPageFaults   *prometheus.Desc
	containerSwapAvailableBytes      *prometheus.Desc
	containerSwapUsageBytes          *prometheus.Desc
	containerAcceleratorMemoryTotal  *prometheus.Desc
	containerAcceleratorMemoryUsed   *prometheus.Desc
	containerAcceleratorDutyCycle    *prometheus.Desc
}

func NewScraper(logger *zap.Logger, targetIP string, tokenPath string, timeout time.Duration) *Scraper {
	return &Scraper{
		tokenPath: tokenPath,
		timeout:   timeout,
		targetIP:  targetIP,
		logger:    logger.With(zap.String("component", "scraper")),
		containerRootFsUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_fs", "usage_bytes"),
			"Disk used in bytes",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerRootFsAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_fs", "limit_bytes"),
			"Capacity of container disk",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerRootFsInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_fs", "inodes_free"),
			"Number of inodes free in container fs",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerRootFsInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_fs", "inodes"),
			"Number of inodes in container fs",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerRootFsInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_fs", "inodes_used"),
			"Number of inodes used in container fs",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerLogsUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_logs", "usage_bytes"),
			"Logs space used in bytes",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerLogsAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_logs", "limit_bytes"),
			"Capacity of container log space",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerLogsInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_logs", "inodes_free"),
			"Number of inodes free in container log space",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerLogsInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_logs", "inodes"),
			"Number of inodes in container log space",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerLogsInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_logs", "inodes_used"),
			"Number of inodes used in container log space",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerCPUUsageNanoCores: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_cpu", "usage_nano_cores"),
			"CPU usage in nanocores",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerCPUUsageCoreNanoSeconds: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_cpu", "usage_core_nano_seconds"),
			"CPU nanoseconds used",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerMemoryAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_memory", "available_bytes"),
			"available bytes in container memory",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerMemoryUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_memory", "usage_bytes"),
			"Used bytes in container memory",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerMemoryWorkingSetBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_memory", "working_set_bytes"),
			"working set bytes in container memory",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerMemoryRSSBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_memory", "rss_bytes"),
			"rss bytes in container memory",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerMemoryPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_memory", "page_faults"),
			"Page faults in container memory",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerMemoryMajorPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_memory", "major_page_faults"),
			"Major page faults in container memory",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerSwapAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_swap", "available_bytes"),
			"Available bytes in container's swap storage",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerSwapUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_swap", "usage_bytes"),
			"Used bytes in container's swap storage",
			[]string{"node", "namespace", "pod", "container"},
			nil),
		containerAcceleratorMemoryTotal: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_accelerator", "memory_total"),
			"Total memory in container's accelerator",
			[]string{"node", "namespace", "pod", "container", "id", "model", "make"},
			nil),
		containerAcceleratorMemoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_accelerator", "memory_used"),
			"Memory used in container's accelerator",
			[]string{"node", "namespace", "pod", "container", "id", "model", "make"},
			nil),
		containerAcceleratorDutyCycle: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "container_accelerator", "duty_cycle"),
			"Percentage of time over which accelerator was allocated",
			[]string{"node", "namespace", "pod", "container", "id", "model", "make"},
			nil),

		podCPUUsageNanoCores: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_cpu", "usage_nano_cores"),
			"CPU usage in nanocores",
			[]string{"node", "namespace", "pod"},
			nil),
		podCPUUsageCoreNanoSeconds: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_cpu", "usage_core_nano_seconds"),
			"CPU nanoseconds used",
			[]string{"node", "namespace", "pod"},
			nil),
		podMemoryAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_memory", "available_bytes"),
			"available bytes in pod memory",
			[]string{"node", "namespace", "pod"},
			nil),
		podMemoryUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_memory", "usage_bytes"),
			"Used bytes in pod memory",
			[]string{"node", "namespace", "pod"},
			nil),
		podMemoryWorkingSetBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_memory", "working_set_bytes"),
			"working set bytes in pod memory",
			[]string{"node", "namespace", "pod"},
			nil),
		podMemoryRSSBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_memory", "rss_bytes"),
			"rss bytes in pod memory",
			[]string{"node", "namespace", "pod"},
			nil),
		podMemoryPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_memory", "page_faults"),
			"Page faults in pod memory",
			[]string{"node", "namespace", "pod"},
			nil),
		podMemoryMajorPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_memory", "major_page_faults"),
			"Major page faults in pod memory",
			[]string{"node", "namespace", "pod"},
			nil),
		podSwapAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_swap", "available_bytes"),
			"Available bytes in pod's swap storage",
			[]string{"node", "namespace", "pod"},
			nil),
		podSwapUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_swap", "usage_bytes"),
			"Used bytes in pod's swap storage",
			[]string{"node", "namespace", "pod"},
			nil),
		podEphemeralStorageUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_ephemeral_storage", "usage_bytes"),
			"Amount of bytes used in pod's ephemeral storage",
			[]string{"node", "namespace", "pod"},
			nil),
		podEphemeralStorageAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_ephemeral_storage", "limit_bytes"),
			"Capacity of pod's ephemeral storage in bytes",
			[]string{"node", "namespace", "pod"},
			nil),
		podEphemeralStorageInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_ephemeral_storage", "inodes_free"),
			"Number of inodes free in pod's ephemeral storage",
			[]string{"node", "namespace", "pod"},
			nil),
		podEphemeralStorageInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_ephemeral_storage", "inodes"),
			"Number of inodes in pod's ephemeral storage",
			[]string{"node", "namespace", "pod"},
			nil),
		podEphemeralStorageInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_ephemeral_storage", "inodes_used"),
			"Number of inodes used in pod's ephemeral storage",
			[]string{"node", "namespace", "pod"},
			nil),
		podVolumeUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_volume", "usage_bytes"),
			"Pod volume used in bytes",
			[]string{"node", "namespace", "pod", "volume_name"},
			nil),
		podVolumeAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_volume", "limit_bytes"),
			"Capacity of pod volume in bytes",
			[]string{"node", "namespace", "pod", "volume_name"},
			nil),
		podVolumeInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_volume", "inodes_free"),
			"Number of inodes free in pod volume",
			[]string{"node", "namespace", "pod", "volume_name"},
			nil),
		podVolumeInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_volume", "inodes"),
			"Number of inodes in pod volume",
			[]string{"node", "namespace", "pod", "volume_name"},
			nil),
		podVolumeInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_volume", "inodes_used"),
			"Number of inodes used in pod volume",
			[]string{"node", "namespace", "pod", "volume_name"},
			nil),
		podVolumeHealthStatus: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_volume", "health_status"),
			"Health status of pod volume",
			[]string{"node", "namespace", "pod", "volume_name"},
			nil),
		podInterfaceRxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_interface", "rx_bytes"),
			"Cumulative count of receive bytes",
			[]string{"node", "namespace", "pod", "name"},
			nil),
		podInterfaceRxErrors: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_interface", "rx_errors"),
			"Cumulative count of receive errors",
			[]string{"node", "namespace", "pod", "name"},
			nil),
		podInterfaceTxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_interface", "tx_bytes"),
			"Cumulative count of transmit bytes",
			[]string{"node", "namespace", "pod", "name"},
			nil),
		podInterfaceTxErrors: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod_interface", "tx_errors"),
			"Cumulative count of transmit errors",
			[]string{"node", "namespace", "pod", "name"},
			nil),
		podProcessCount: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "pod", "process_count"),
			"Count of process in pod",
			[]string{"node", "namespace", "pod"},
			nil),
		nodeFsUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_fs", "usage_bytes"),
			"Disk used in bytes",
			[]string{"node"},
			nil),
		nodeFsAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_fs", "limit_bytes"),
			"Capacity of container disk",
			[]string{"node"},
			nil),
		nodeFsInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_fs", "inodes_free"),
			"Number of inodes free in node fs",
			[]string{"node"},
			nil),
		nodeFsInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_fs", "inodes"),
			"Number of inodes in node fs",
			[]string{"node"},
			nil),
		nodeFsInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_fs", "inodes_used"),
			"Number of inodes used in node fs",
			[]string{"node"},
			nil),
		nodeRuntimeImageFsUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_image_fs", "usage_bytes"),
			"Usage of node runtime image fs in bytes",
			[]string{"node"},
			nil),
		nodeRuntimeImageFsAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_image_fs", "limit_bytes"),
			"Capacity of node runtime image fs in bytes",
			[]string{"node"},
			nil),
		nodeRuntimeImageFsInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_image_fs", "inodes_free"),
			"Inodes free in node's runtime image fs",
			[]string{"node"},
			nil),
		nodeRuntimeImageFsInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_image_fs", "inodes"),
			"Inodes in node's runtime image fs",
			[]string{"node"},
			nil),
		nodeRuntimeImageFsInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_image_fs", "inodes_used"),
			"Inodes used in node's runtime image fs",
			[]string{"node"},
			nil),
		nodeRuntimeContainerFsUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_container_fs", "usage_bytes"),
			"Usage of node runtime container's writeable layer in bytes",
			[]string{"node"},
			nil),
		nodeRuntimeContainerFsAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_container_fs", "limit_bytes"),
			"Capacity of node runtime container's writeable layer in bytes",
			[]string{"node"},
			nil),
		nodeRuntimeContainerFsInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_container_fs", "inodes_free"),
			"Count of free Inodes in node runtime container fs",
			[]string{"node"},
			nil),
		nodeRuntimeContainerFsInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_container_fs", "inodes"),
			"Total inodes in node's runtime container fs",
			[]string{"node"},
			nil),
		nodeRuntimeContainerFsInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_runtime_container_fs", "inodes_used"),
			"Count of inodes being used",
			[]string{"node"},
			nil),
		nodeCPUUsageNanoCores: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_cpu", "usage_nano_cores"),
			"CPU usage in nanocores",
			[]string{"node"},
			nil),
		nodeCPUUsageCoreNanoSeconds: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_cpu", "usage_core_nano_seconds"),
			"CPU nanoseconds used",
			[]string{"node"},
			nil),
		nodeMemoryAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_memory", "available_bytes"),
			"available bytes in node memory",
			[]string{"node"},
			nil),
		nodeMemoryUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_memory", "usage_bytes"),
			"Used bytes in node memory",
			[]string{"node"},
			nil),
		nodeMemoryWorkingSetBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_memory", "working_set_bytes"),
			"working set bytes in node memory",
			[]string{"node"},
			nil),
		nodeMemoryRSSBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_memory", "rss_bytes"),
			"rss bytes in node memory",
			[]string{"node"},
			nil),
		nodeMemoryPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_memory", "page_faults"),
			"Page faults in node memory",
			[]string{"node"},
			nil),
		nodeMemoryMajorPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_memory", "major_page_faults"),
			"Major page faults in node memory",
			[]string{"node"},
			nil),
		nodeSwapAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_swap", "available_bytes"),
			"Available bytes in node's swap storage",
			[]string{"node"},
			nil),
		nodeSwapUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_swap", "usage_bytes"),
			"Used bytes in node's swap storage",
			[]string{"node"},
			nil),
		nodeRLimitMaxPID: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_rlimit", "max_pid"),
			"Maximum PID",
			[]string{"node"},
			nil),
		nodeRLimitNumOfRunningProcess: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_rlimit", "num_of_running_process"),
			"Number of running process in node",
			[]string{"node"},
			nil),
		nodeInterfaceRxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_interface", "rx_bytes"),
			"Cumulative count of receive bytes",
			[]string{"node", "name"},
			nil),
		nodeInterfaceRxErrors: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_interface", "rx_errors"),
			"Cumulative count of receive errors",
			[]string{"node", "name"},
			nil),
		nodeInterfaceTxBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_interface", "tx_bytes"),
			"Cumulative count of transmit bytes",
			[]string{"node", "name"},
			nil),
		nodeInterfaceTxErrors: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_interface", "tx_errors"),
			"Cumulative count of transmit errors",
			[]string{"node", "name"},
			nil),
		nodeSystemContainerRootFsUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_fs", "usage_bytes"),
			"Disk used in bytes",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerRootFsAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_fs", "limit_bytes"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerRootFsInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_fs", "inodes_free"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerRootFsInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_fs", "inodes"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerRootFsInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_fs", "inodes_used"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerLogsUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_logs", "usage_bytes"),
			"Disk used in bytes",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerLogsAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_logs", "limit_bytes"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerLogsInodesFree: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_logs", "inodes_free"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerLogsInodes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_logs", "inodes"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerLogsInodesUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_logs", "inodes_used"),
			"Capacity of nodeSystemContainer disk",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerCPUUsageNanoCores: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_cpu", "usage_nano_cores"),
			"CPU usage in nanocores",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerCPUUsageCoreNanoSeconds: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_cpu", "usage_core_nano_seconds"),
			"CPU usage in core nanoseconds",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerMemoryAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_memory", "available_bytes"),
			"available bytes in nodeSystemContainer memory",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerMemoryUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_memory", "usage_bytes"),
			"Used bytes in nodeSystemContainer memory",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerMemoryWorkingSetBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_memory", "working_set_bytes"),
			"working set bytes in nodeSystemContainer memory",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerMemoryRSSBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_memory", "rss_bytes"),
			"rss bytes in nodeSystemContainer memory",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerMemoryPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_memory", "page_faults"),
			"Page faults in nodeSystemContainer memory",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerMemoryMajorPageFaults: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_memory", "major_page_faults"),
			"Major page faults in nodeSystemContainer memory",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerSwapAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_swap", "available_bytes"),
			"Available bytes in nodeSystemContainer's swap storage",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerSwapUsageBytes: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_swap", "usage_bytes"),
			"Used bytes in nodeSystemContainer's swap storage",
			[]string{"node", "container"},
			nil),
		nodeSystemContainerAcceleratorMemoryTotal: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_accelerator", "memory_total"),
			"Total memory in nodeSystemContainer's accelerator",
			[]string{"node", "container", "id", "model", "make"},
			nil),
		nodeSystemContainerAcceleratorMemoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_accelerator", "memory_used"),
			"Memory used in nodeSystemContainer's accelerator",
			[]string{"node", "container", "id", "model", "make"},
			nil),
		nodeSystemContainerAcceleratorDutyCycle: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary", "node_system_container_accelerator", "duty_cycle"),
			"Percentage of time over which accelerator was allocated",
			[]string{"node", "container", "id", "model", "make"},
			nil),
		errors: prometheus.NewDesc(
			prometheus.BuildFQName("kubelet_summary_exporter", "", "errors"),
			"Errors scraping kubelet stats summary",
			[]string{"type"},
			nil),
	}
}

func (s *Scraper) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.errors

	ch <- s.nodeFsUsedBytes
	ch <- s.nodeFsAvailableBytes
	ch <- s.nodeFsInodesFree
	ch <- s.nodeFsInodes
	ch <- s.nodeFsInodesUsed
	ch <- s.nodeRuntimeImageFsUsedBytes
	ch <- s.nodeRuntimeImageFsAvailableBytes
	ch <- s.nodeRuntimeImageFsInodesFree
	ch <- s.nodeRuntimeImageFsInodes
	ch <- s.nodeRuntimeImageFsInodesUsed
	ch <- s.nodeRuntimeContainerFsUsedBytes
	ch <- s.nodeRuntimeContainerFsAvailableBytes
	ch <- s.nodeRuntimeContainerFsInodesFree
	ch <- s.nodeRuntimeContainerFsInodes
	ch <- s.nodeRuntimeContainerFsInodesUsed
	ch <- s.nodeCPUUsageNanoCores
	ch <- s.nodeCPUUsageCoreNanoSeconds
	ch <- s.nodeMemoryAvailableBytes
	ch <- s.nodeMemoryUsageBytes
	ch <- s.nodeMemoryWorkingSetBytes
	ch <- s.nodeMemoryRSSBytes
	ch <- s.nodeMemoryPageFaults
	ch <- s.nodeMemoryMajorPageFaults
	ch <- s.nodeSwapAvailableBytes
	ch <- s.nodeSwapUsageBytes
	ch <- s.nodeRLimitMaxPID
	ch <- s.nodeRLimitNumOfRunningProcess
	ch <- s.nodeInterfaceRxBytes
	ch <- s.nodeInterfaceRxErrors
	ch <- s.nodeInterfaceTxBytes
	ch <- s.nodeInterfaceTxErrors
	ch <- s.nodeSystemContainerRootFsUsedBytes
	ch <- s.nodeSystemContainerRootFsAvailableBytes
	ch <- s.nodeSystemContainerRootFsInodesFree
	ch <- s.nodeSystemContainerRootFsInodes
	ch <- s.nodeSystemContainerRootFsInodesUsed
	ch <- s.nodeSystemContainerLogsUsedBytes
	ch <- s.nodeSystemContainerLogsAvailableBytes
	ch <- s.nodeSystemContainerLogsInodesFree
	ch <- s.nodeSystemContainerLogsInodes
	ch <- s.nodeSystemContainerLogsInodesUsed
	ch <- s.nodeSystemContainerCPUUsageNanoCores
	ch <- s.nodeSystemContainerCPUUsageCoreNanoSeconds
	ch <- s.nodeSystemContainerMemoryAvailableBytes
	ch <- s.nodeSystemContainerMemoryUsageBytes
	ch <- s.nodeSystemContainerMemoryWorkingSetBytes
	ch <- s.nodeSystemContainerMemoryRSSBytes
	ch <- s.nodeSystemContainerMemoryPageFaults
	ch <- s.nodeSystemContainerMemoryMajorPageFaults
	ch <- s.nodeSystemContainerSwapAvailableBytes
	ch <- s.nodeSystemContainerSwapUsageBytes
	ch <- s.nodeSystemContainerAcceleratorMemoryTotal
	ch <- s.nodeSystemContainerAcceleratorMemoryUsed
	ch <- s.nodeSystemContainerAcceleratorDutyCycle

	ch <- s.podCPUUsageNanoCores
	ch <- s.podCPUUsageCoreNanoSeconds
	ch <- s.podMemoryAvailableBytes
	ch <- s.podMemoryUsageBytes
	ch <- s.podMemoryWorkingSetBytes
	ch <- s.podMemoryRSSBytes
	ch <- s.podMemoryPageFaults
	ch <- s.podMemoryMajorPageFaults
	ch <- s.podSwapAvailableBytes
	ch <- s.podSwapUsageBytes
	ch <- s.podInterfaceRxBytes
	ch <- s.podInterfaceRxErrors
	ch <- s.podInterfaceTxBytes
	ch <- s.podInterfaceTxErrors
	ch <- s.podEphemeralStorageUsedBytes
	ch <- s.podEphemeralStorageAvailableBytes
	ch <- s.podEphemeralStorageInodesFree
	ch <- s.podEphemeralStorageInodes
	ch <- s.podEphemeralStorageInodesUsed
	ch <- s.podVolumeUsedBytes
	ch <- s.podVolumeAvailableBytes
	ch <- s.podVolumeInodesFree
	ch <- s.podVolumeInodes
	ch <- s.podVolumeInodesUsed
	ch <- s.podVolumeHealthStatus
	ch <- s.podProcessCount

	ch <- s.containerRootFsUsedBytes
	ch <- s.containerRootFsAvailableBytes
	ch <- s.containerRootFsInodesFree
	ch <- s.containerRootFsInodes
	ch <- s.containerRootFsInodesUsed
	ch <- s.containerLogsUsedBytes
	ch <- s.containerLogsAvailableBytes
	ch <- s.containerLogsInodesFree
	ch <- s.containerLogsInodes
	ch <- s.containerLogsInodesUsed
	ch <- s.containerCPUUsageNanoCores
	ch <- s.containerCPUUsageCoreNanoSeconds
	ch <- s.containerMemoryAvailableBytes
	ch <- s.containerMemoryUsageBytes
	ch <- s.containerMemoryWorkingSetBytes
	ch <- s.containerMemoryRSSBytes
	ch <- s.containerMemoryPageFaults
	ch <- s.containerMemoryMajorPageFaults
	ch <- s.containerSwapAvailableBytes
	ch <- s.containerSwapUsageBytes
	ch <- s.containerAcceleratorMemoryTotal
	ch <- s.containerAcceleratorMemoryUsed
	ch <- s.containerAcceleratorDutyCycle
}

func (s *Scraper) Collect(ch chan<- prometheus.Metric) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s:10250/stats/summary", s.targetIP), nil)
	if err != nil {
		s.logger.Error("failed to create request", zap.Error(err))
		return
	}

	token, err := os.ReadFile(s.tokenPath)
	if err != nil {
		s.logger.Fatal("unable to load specified token", zap.String("file", s.tokenPath), zap.Error(err))
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{
		Timeout: s.timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec //See https://nvbugspro.nvidia.com/bug/4474467
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		s.errCnt++
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"request error",
		)
		s.logger.Warn("failed to make request to stats/summary", zap.Error(err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		s.errCnt++
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"status error",
		)
		s.logger.Warn("got unexpected status for stats/summary", zap.String("status", resp.Status))
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.errCnt++
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"read body error",
		)
		s.logger.Error("failed to read body", zap.Error(err))
		return
	}

	summary, err := s.parse(body)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			s.errors,
			prometheus.CounterValue,
			s.errCnt,
			"parse body error",
		)
		s.logger.Error("failed to parse body", zap.Error(err))
		return
	}
	node := summary.Node
	nodeName := node.NodeName
	for _, nodeSystemContainer := range node.SystemContainers {
		if nodeSystemContainer.Rootfs != nil {
			s.pushMetrics(ch, s.nodeSystemContainerRootFsUsedBytes, nodeSystemContainer.Rootfs.UsedBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerRootFsAvailableBytes, nodeSystemContainer.Rootfs.CapacityBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerRootFsInodes, nodeSystemContainer.Rootfs.Inodes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerRootFsInodesFree, nodeSystemContainer.Rootfs.InodesFree, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerRootFsInodesUsed, nodeSystemContainer.Rootfs.InodesUsed, nodeName, nodeSystemContainer.Name)
		}

		if nodeSystemContainer.Logs != nil {
			s.pushMetrics(ch, s.nodeSystemContainerLogsUsedBytes, nodeSystemContainer.Logs.UsedBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerLogsAvailableBytes, nodeSystemContainer.Logs.CapacityBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerLogsInodes, nodeSystemContainer.Logs.Inodes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerLogsInodesFree, nodeSystemContainer.Logs.InodesFree, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerLogsInodesUsed, nodeSystemContainer.Logs.InodesUsed, nodeName, nodeSystemContainer.Name)
		}

		if nodeSystemContainer.CPU != nil {
			s.pushMetrics(ch, s.nodeSystemContainerCPUUsageNanoCores, nodeSystemContainer.CPU.UsageNanoCores, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerCPUUsageCoreNanoSeconds, nodeSystemContainer.CPU.UsageCoreNanoSeconds, nodeName, nodeSystemContainer.Name)
		}

		if nodeSystemContainer.Memory != nil {
			fmt.Println(nodeSystemContainer.Memory, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerMemoryAvailableBytes, nodeSystemContainer.Memory.AvailableBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerMemoryUsageBytes, nodeSystemContainer.Memory.UsageBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerMemoryWorkingSetBytes, nodeSystemContainer.Memory.WorkingSetBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerMemoryRSSBytes, nodeSystemContainer.Memory.RSSBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerMemoryPageFaults, nodeSystemContainer.Memory.PageFaults, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerMemoryMajorPageFaults, nodeSystemContainer.Memory.MajorPageFaults, nodeName, nodeSystemContainer.Name)
		}

		if nodeSystemContainer.Swap != nil {
			s.pushMetrics(ch, s.nodeSystemContainerSwapAvailableBytes, nodeSystemContainer.Swap.SwapAvailableBytes, nodeName, nodeSystemContainer.Name)
			s.pushMetrics(ch, s.nodeSystemContainerSwapUsageBytes, nodeSystemContainer.Swap.SwapUsageBytes, nodeName, nodeSystemContainer.Name)
		}

		for _, accelerator := range nodeSystemContainer.Accelerators {
			s.pushMetrics(ch, s.nodeSystemContainerAcceleratorMemoryUsed, &accelerator.MemoryUsed, nodeName, nodeSystemContainer.Name, accelerator.ID, accelerator.Model, accelerator.Make)
			s.pushMetrics(ch, s.nodeSystemContainerAcceleratorMemoryTotal, &accelerator.MemoryTotal, nodeName, nodeSystemContainer.Name, accelerator.ID, accelerator.Model, accelerator.Make)
			s.pushMetrics(ch, s.nodeSystemContainerAcceleratorDutyCycle, &accelerator.DutyCycle, nodeName, nodeSystemContainer.Name, accelerator.ID, accelerator.Model, accelerator.Make)
		}
	}

	nodeFs := node.Fs
	if nodeFs != nil {
		s.pushMetrics(ch, s.nodeFsUsedBytes, nodeFs.UsedBytes, nodeName)
		s.pushMetrics(ch, s.nodeFsAvailableBytes, nodeFs.CapacityBytes, nodeName)
		s.pushMetrics(ch, s.nodeFsInodes, nodeFs.Inodes, nodeName)
		s.pushMetrics(ch, s.nodeFsInodesFree, nodeFs.InodesFree, nodeName)
		s.pushMetrics(ch, s.nodeFsInodesUsed, nodeFs.InodesUsed, nodeName)
	}

	nodeRuntimeImageFs := node.Runtime.ImageFs
	if nodeRuntimeImageFs != nil {
		s.pushMetrics(ch, s.nodeRuntimeImageFsUsedBytes, nodeRuntimeImageFs.UsedBytes, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeImageFsAvailableBytes, nodeRuntimeImageFs.CapacityBytes, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeImageFsInodes, nodeRuntimeImageFs.Inodes, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeImageFsInodesFree, nodeRuntimeImageFs.InodesFree, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeImageFsInodesUsed, nodeRuntimeImageFs.InodesUsed, nodeName)
	}

	nodeRuntimeContainerFs := node.Runtime.ContainerFs
	if nodeRuntimeContainerFs != nil {
		s.pushMetrics(ch, s.nodeRuntimeContainerFsUsedBytes, nodeRuntimeContainerFs.UsedBytes, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeContainerFsAvailableBytes, nodeRuntimeContainerFs.CapacityBytes, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeContainerFsInodes, nodeRuntimeContainerFs.Inodes, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeContainerFsInodesFree, nodeRuntimeContainerFs.InodesFree, nodeName)
		s.pushMetrics(ch, s.nodeRuntimeContainerFsInodesUsed, nodeRuntimeContainerFs.InodesUsed, nodeName)
	}

	if node.CPU != nil {
		s.pushMetrics(ch, s.nodeCPUUsageNanoCores, node.CPU.UsageNanoCores, nodeName)
		s.pushMetrics(ch, s.nodeCPUUsageCoreNanoSeconds, node.CPU.UsageCoreNanoSeconds, nodeName)
	}

	if node.Memory != nil {
		s.pushMetrics(ch, s.nodeMemoryAvailableBytes, node.Memory.AvailableBytes, nodeName)
		s.pushMetrics(ch, s.nodeMemoryUsageBytes, node.Memory.UsageBytes, nodeName)
		s.pushMetrics(ch, s.nodeMemoryWorkingSetBytes, node.Memory.WorkingSetBytes, nodeName)
		s.pushMetrics(ch, s.nodeMemoryRSSBytes, node.Memory.RSSBytes, nodeName)
		s.pushMetrics(ch, s.nodeMemoryPageFaults, node.Memory.PageFaults, nodeName)
	}

	if node.Swap != nil {
		s.pushMetrics(ch, s.nodeSwapAvailableBytes, node.Swap.SwapAvailableBytes, nodeName)
		s.pushMetrics(ch, s.nodeSwapUsageBytes, node.Swap.SwapUsageBytes, nodeName)
	}

	if node.Rlimit != nil {
		if node.Rlimit.MaxPID != nil {
			ch <- prometheus.MustNewConstMetric(
				s.nodeRLimitMaxPID,
				prometheus.GaugeValue,
				float64(*node.Rlimit.MaxPID),
				nodeName,
			)
		}
		if node.Rlimit.NumOfRunningProcesses != nil {
			ch <- prometheus.MustNewConstMetric(
				s.nodeRLimitNumOfRunningProcess,
				prometheus.GaugeValue,
				float64(*node.Rlimit.NumOfRunningProcesses),
				nodeName,
			)
		}
	}

	if node.Network != nil {
		for _, interfaceStats := range node.Network.Interfaces {
			interfaceName := interfaceStats.Name
			s.pushMetrics(ch, s.nodeInterfaceRxBytes, interfaceStats.RxBytes, nodeName, interfaceName)
			s.pushMetrics(ch, s.nodeInterfaceRxErrors, interfaceStats.RxErrors, nodeName, interfaceName)
			s.pushMetrics(ch, s.nodeInterfaceTxBytes, interfaceStats.TxBytes, nodeName, interfaceName)
			s.pushMetrics(ch, s.nodeInterfaceTxErrors, interfaceStats.TxErrors, nodeName, interfaceName)
		}
	}

	for _, pod := range summary.Pods {
		podName := pod.PodRef.Name
		namespace := pod.PodRef.Namespace
		if pod.CPU != nil {
			s.pushMetrics(ch, s.podCPUUsageNanoCores, pod.CPU.UsageNanoCores, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podCPUUsageCoreNanoSeconds, pod.CPU.UsageCoreNanoSeconds, nodeName, namespace, podName)
		}

		if pod.Memory != nil {
			s.pushMetrics(ch, s.podMemoryAvailableBytes, pod.Memory.AvailableBytes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podMemoryUsageBytes, pod.Memory.UsageBytes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podMemoryWorkingSetBytes, pod.Memory.WorkingSetBytes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podMemoryRSSBytes, pod.Memory.RSSBytes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podMemoryPageFaults, pod.Memory.PageFaults, nodeName, namespace, podName)
		}

		if pod.Swap != nil {
			s.pushMetrics(ch, s.podSwapAvailableBytes, pod.Swap.SwapAvailableBytes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podSwapUsageBytes, pod.Swap.SwapUsageBytes, nodeName, namespace, podName)
		}

		if pod.EphemeralStorage != nil {
			s.pushMetrics(ch, s.podEphemeralStorageUsedBytes, pod.EphemeralStorage.UsedBytes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podEphemeralStorageAvailableBytes, pod.EphemeralStorage.CapacityBytes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podEphemeralStorageInodes, pod.EphemeralStorage.Inodes, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podEphemeralStorageInodesFree, pod.EphemeralStorage.InodesFree, nodeName, namespace, podName)
			s.pushMetrics(ch, s.podEphemeralStorageInodesUsed, pod.EphemeralStorage.InodesUsed, nodeName, namespace, podName)
		}

		if pod.ProcessStats != nil {
			s.pushMetrics(ch, s.podProcessCount, pod.ProcessStats.ProcessCount, nodeName, namespace, podName)
		}

		for _, podVolume := range pod.VolumeStats {
			s.pushMetrics(ch, s.podVolumeUsedBytes, podVolume.FsStats.UsedBytes, nodeName, namespace, podName, podVolume.Name)
			s.pushMetrics(ch, s.podVolumeAvailableBytes, podVolume.FsStats.CapacityBytes, nodeName, namespace, podName, podVolume.Name)
			s.pushMetrics(ch, s.podVolumeInodes, podVolume.FsStats.Inodes, nodeName, namespace, podName, podVolume.Name)
			s.pushMetrics(ch, s.podVolumeInodesFree, podVolume.FsStats.InodesFree, nodeName, namespace, podName, podVolume.Name)
			s.pushMetrics(ch, s.podVolumeInodesUsed, podVolume.FsStats.InodesUsed, nodeName, namespace, podName, podVolume.Name)

			if podVolume.VolumeHealthStats != nil {
				var podVolumeHealthStatus uint64 = 0
				if podVolume.VolumeHealthStats.Abnormal {
					podVolumeHealthStatus = 1
				}
				s.pushMetrics(ch, s.podVolumeHealthStatus, &podVolumeHealthStatus, nodeName, namespace, podName, podVolume.Name)
			}
		}

		if pod.Network != nil {
			for _, interfaceStats := range pod.Network.Interfaces {
				interfaceName := interfaceStats.Name
				s.pushMetrics(ch, s.podInterfaceRxBytes, interfaceStats.RxBytes, nodeName, namespace, podName, interfaceName)
				s.pushMetrics(ch, s.podInterfaceRxErrors, interfaceStats.RxErrors, nodeName, namespace, podName, interfaceName)
				s.pushMetrics(ch, s.podInterfaceTxBytes, interfaceStats.TxBytes, nodeName, namespace, podName, interfaceName)
				s.pushMetrics(ch, s.podInterfaceTxErrors, interfaceStats.TxErrors, nodeName, namespace, podName, interfaceName)
			}
		}
		for _, container := range pod.Containers {
			if container.Rootfs != nil {
				s.pushMetrics(ch, s.containerRootFsUsedBytes, container.Rootfs.UsedBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerRootFsAvailableBytes, container.Rootfs.CapacityBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerRootFsInodes, container.Rootfs.Inodes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerRootFsInodesFree, container.Rootfs.InodesFree, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerRootFsInodesUsed, container.Rootfs.InodesUsed, nodeName, namespace, podName, container.Name)
			}

			if container.Logs != nil {
				s.pushMetrics(ch, s.containerLogsUsedBytes, container.Logs.UsedBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerLogsAvailableBytes, container.Logs.CapacityBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerLogsInodes, container.Logs.Inodes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerLogsInodesFree, container.Logs.InodesFree, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerLogsInodesUsed, container.Logs.InodesUsed, nodeName, namespace, podName, container.Name)
			}

			if container.CPU != nil {
				s.pushMetrics(ch, s.containerCPUUsageNanoCores, container.CPU.UsageNanoCores, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerCPUUsageCoreNanoSeconds, container.CPU.UsageCoreNanoSeconds, nodeName, namespace, podName, container.Name)
			}

			if container.Memory != nil {
				s.pushMetrics(ch, s.containerMemoryAvailableBytes, container.Memory.AvailableBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerMemoryUsageBytes, container.Memory.UsageBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerMemoryWorkingSetBytes, container.Memory.WorkingSetBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerMemoryRSSBytes, container.Memory.RSSBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerMemoryPageFaults, container.Memory.PageFaults, nodeName, namespace, podName, container.Name)
			}

			if container.Swap != nil {
				s.pushMetrics(ch, s.containerSwapAvailableBytes, container.Swap.SwapAvailableBytes, nodeName, namespace, podName, container.Name)
				s.pushMetrics(ch, s.containerSwapUsageBytes, container.Swap.SwapUsageBytes, nodeName, namespace, podName, container.Name)
			}

			for _, accelerator := range container.Accelerators {
				s.pushMetrics(ch, s.containerAcceleratorMemoryUsed, &accelerator.MemoryUsed, nodeName, podName, namespace, container.Name, accelerator.ID, accelerator.Model, accelerator.Make)
				s.pushMetrics(ch, s.containerAcceleratorMemoryTotal, &accelerator.MemoryTotal, nodeName, podName, namespace, container.Name, accelerator.ID, accelerator.Model, accelerator.Make)
				s.pushMetrics(ch, s.containerAcceleratorDutyCycle, &accelerator.DutyCycle, nodeName, podName, namespace, container.Name, accelerator.ID, accelerator.Model, accelerator.Make)
			}
		}
	}
}

func (s *Scraper) parse(body []byte) (*statsapi.Summary, error) {
	var summary statsapi.Summary
	err := json.Unmarshal(body, &summary)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (s *Scraper) pushMetrics(ch chan<- prometheus.Metric, metric *prometheus.Desc, value *uint64, labelValues ...string) {
	if value != nil {
		ch <- prometheus.MustNewConstMetric(
			metric,
			prometheus.GaugeValue,
			float64(*value),
			labelValues...,
		)
	}
}
