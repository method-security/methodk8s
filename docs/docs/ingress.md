# Ingress

The `methodk8s ingress` family of commands provide information about a cluster's ingresses.

## Enumerate

The enumerate command will gather information about all of the ingresses that the provided context have access to. This includes gateway objects as well. If you only want to enumerate gateway objects please see flag options below

### Usage

```bash
methodk8s ingress enumerate
```

### Help Text

```bash
$ methodk8s ingress enumerate -h
Enumerate Ingress objects

Usage:
  methodk8s ingress enumerate [flags]

Flags:
  -h, --help            help for enumerate
      --type strings   List the types to emumerate (ie. --type ingress --type gateway --type loadbalancer)

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
