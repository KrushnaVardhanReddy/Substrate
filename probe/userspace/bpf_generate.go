package userspace

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpfel -cc clang bpf ../ebpf/http_capture.c -- -I../headers -I/usr/include -I/usr/include/x86_64-linux-gnu
