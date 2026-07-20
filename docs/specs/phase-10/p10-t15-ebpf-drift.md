# P10-T15: Zero-Latency Drift Detection (eBPF)

## Objective
Extend Substrate's runtime drift detection capabilities with a zero-latency `cilium/ebpf` kernel probe. This is designed for high-throughput enterprise environments where sidecar proxies or application-level middleware introduce unacceptable overhead. The eBPF probe will capture network traffic at the kernel level and compare live payloads against the OpenAPI spec in real-time.

## Core Features
1. **Kernel-Level Packet Capture (eBPF):**
   Utilize `cilium/ebpf` to hook into kernel networking constructs (e.g., TC or socket filters). Capture ingress/egress HTTP/gRPC traffic directed at API pods without modifying the application code or routing it through a proxy.
   
2. **Real-time Spec Comparison:**
   As packets are reconstructed, parse the request/response payloads in a highly optimized Go user-space application and stream them to Substrate's diff engine to detect undocumented endpoints, missing fields, or data type violations in real-time.

3. **Kubernetes Native Auto-Discovery:**
   Deploy as a DaemonSet. The probe should automatically detect and monitor pods based on Kubernetes labels/annotations (e.g., `substrate.io/monitor: "true"`), dynamically attaching/detaching eBPF programs as pods spin up or down.

4. **Zero-Overhead UX:**
   Provide a pre-packaged Helm chart (`helm install substrate-ebpf`) for frictionless deployment by platform engineering teams.

## Deliverables
- `probe/ebpf/`: The C-based eBPF programs and compiled BPF objects.
- `probe/userspace/`: The Go application leveraging `cilium/ebpf` to load the programs, read perf rings/maps, and send telemetry to the Substrate core.
- `charts/substrate-ebpf/`: Helm chart for deploying the daemonset with RBAC configurations to allow kernel tracing.
- E2E testing using kind (Kubernetes in Docker) to validate traffic capture and label-based pod discovery.
