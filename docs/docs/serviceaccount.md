# ServiceAccount

The `methodk8s serviceaccount` family of commands provide functionaility to seemlessly configer a Service Account to be used by the method agent to scan your enviroments

## Configure Creds

The `configure creds` command will gather information about your Service Account configuration and print it to the command line. Use this command to gather the required data needed to create a K8s Credential in the Method platform
### Usage

```bash
methodk8s serviceaccount configure creds 
```

### Help Text

```bash
$ methodk8s serviceaccount configure creds -h

Usage:
  methodk8s configure creds [flags]

Flags:
  -h, --help                help for creds
      --namespace string    Set the namespace for the Service Account and Secret (default "default")
      --secretname string   The name of the secret to use for authentication (default "method-sa-secret")

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

## Configure Apply

The configure apply command will generate the required .yaml files needed to setup the Service Account in your cluster. The command defaults to 'dry run' where it prints the yaml files to the console.

### Usage

```bash
methodk8s serviceaccount configure apply 
```

### Help Text

```bash
$ methodk8s serviceaccount configure apply  -h
Create a service account in your k8s cluster

Usage:
  methodk8s serviceaccount configure apply [flags]

Flags:
  -h, --help               help for apply
      --namespace string   Set the namespace for the Service Account and Secret (default "default")
      --run                Apply the Service Account yamls (defaults to false)

Global Flags:
  -a, --cert string          Base64 encoded ca certificate
  -c, --context string       Cluster context (ie. minikube)
  -o, --output string        Output format (signal, json, yaml). Default value is signal (default "signal")
  -f, --output-file string   Path to output file. If blank, will output to STDOUT
  -p, --path string          Absolute or relative path to the config file (ie. ~/.kube/config)
  -q, --quiet                Suppress output
  -s, --serviceaccount       Set to true if using service account workflow
  -t, --token string         Base64 Service account token
  -u, --url string           Cluster url (ie. mycluster.com)
  -v, --verbose              Verbose output
```
