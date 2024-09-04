# Node

The `methodk8s node` family of commands provide information about an cluster's nodes and containers.

## Enumerate

The enumerate command will gather information about all of the nodes that the provided context have access to.

### Usage

```bash
methodk8s node enumerate
```

### Help Text

```bash
$ methodk8s node enumerate -h
Enumerate Node objects

Usage:
  methodk8s node enumerate [flags]

Flags:
  -h, --help   help for enumerate

Global Flags:
  -a, --cert string          Base64 encoded CA Certificate
  -c, --context string       Cluster context (ie. minikube)
  -o, --output string        Output format (signal, json, yaml). Default value is signal (default "signal")
  -f, --output-file string   Path to output file. If blank, will output to STDOUT
  -p, --path string          Absolute or relative path to the config file (ie. ~/.kube/config)
  -q, --quiet                Suppress output
  -s, --serviceaccount       Set to True if using service account workflow
  -t, --token string         Base64 Service Account token
  -u, --url string           Cluster URL (ie. mycluster.com)
  -v, --verbose              Verbose output
```
