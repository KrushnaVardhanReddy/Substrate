//go:build ignore

#include <linux/bpf.h>
#include "bpf_helpers.h"

char __license[] SEC("license") = "Dual MIT/GPL";

struct bpf_map_def SEC("maps") events = {
    .type = BPF_MAP_TYPE_RINGBUF,
    .max_entries = 1 << 24,
};

SEC("socket")
int socket_http_filter(struct __sk_buff *skb) {
    return 0;
}
