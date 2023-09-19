# Pomaloba

The Poor man's Load Balancer (ie. for bare-metal Kubernetes clusters)

## Quick example

A picture is worth a thousand words:

```{bash}

    client           LB             Worker1        Worker2        Worker3
                (IP:1.2.3.4)     (IP 10.0.1.2)  (IP 10.0.1.3)  (IP 10.0.1.4) 
    ------.          ---            -------        -------        -------
       |              |                |              |            |
       | GET          |                |              |            |
       |------------->| GET            |              |            |
       |              |--------------->|              |            |
       |              |                |              |            |
       |              |       Response |              |            |
       |      Resonse |<---------------|              |            |
       |<-------------|                |              |            |
       |              |                |              |            |
       |              |                |              |            |
       | GET          |                |              |            |
       |------------->| GET            |              |            |
       |              |------------------------------------------->|
       |              |                |              |            |
       |              |                |              |   Response |
       |     Response |<-------------------------------------------|
       |<-------------|                |              |            |
       |              |                |              |            |
       |              |                |              |            |
       
```

Thus to realize the picture above just type:

```{bash}
cat << EOF > sample.json
{
  "pomaloba": 1,
  "app": "MyApp",
  "endpoints": [ 
    {"ip": "1.2.3.4", "port": 1234 }
  ],
  "destinations": [
    {"ip": "10.0.1.2", "port": 31234 },
    {"ip": "10.0.1.3", "port": 31234 },
    {"ip": "10.0.1.4", "port": 31234 },
  ]
}
EOF

pomaloba --input-spec sample.json --output-type iptables-rules
```

.. and run the resulting content in the LB instance.

## Motivation

Let's say you have your own bare-metal Kubernetes cluster running on
your private cloud or even realized using several VM instances merrily
running in your own desktop (or laptop these days).

Everything is fine until you deploy a workload consisting in several
replicas and want to add a Load Balancer to the mix: a service of type
`Load Balancer` won't realize. What do you do now?

Well you could use Services of type `NodePort` ensuring the traffic
will arrive to its destination once you reach the worker nodes but you
still want to have a *single* (IP address, port) endpoint exponsed to
reach your application.

Thus this Poor Man's Load Balancer exists: just provision a small
Linux VM instance with Linux NetFilter enabled and use the Pomaloba
output to generate the necesary DNAT and Masquerading rules to make it
happen.

For the time being it generates output suitable to be used by
`iptables` but supporting `nftables` and Ansible should not be too
complicated once I wrap my head around it.

