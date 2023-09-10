# Pomaloba

The Poor man's Load Balancer (ie. for bare-metal Kubernetes clusters)

## Quick example

```{bash}
cat << EOF > sample.json
{
  "pomaloba": 1,
  "app": "MyApp",
  "endpoints": [ 
    {"ip": "1.2.3.4", "port": 1234 }
  ],
  "destinations": [
    {"ip": "192.168.100.11", "port": 31234 },
    {"ip": "192.168.100.12", "port": 31234 },
    {"ip": "192.168.100.13", "port": 31234 },
    {"ip": "192.168.100.14", "port": 31234 }
  ]
}
EOF

pomaloba --input-spec sample.json
```

## Motivation

Let's say you have your own bare-metal Kubernetes cluster running on
your private cloud or even realized using several VM instances merrily
running in your own desktop (or laptop these days).

Everything is fine until you deploy a workload consisting in several
replicas and want to add a Load Balancer to the mix: A Service of type
`Load Balancer` won't deploy. What do you do now?`

Well you could use Services of type `NodePort` ensuring the traffic
will arrive to its destination once you reach the worker nodes but you
still want to have a single endpoint to reach your application. Then
provision a small VM instance with Linux NetFilter enabled and use
Pomaloba to generate the necesary NAT rules to make it happen.

For the time being it uses `iptables` but supporting `nftables` should
not be so complicated once I wrap my head around it.

