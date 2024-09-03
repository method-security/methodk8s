# methodk8s Documentation

Hello and welcome to the methodk8s documentation. While we always want to provide the most comprehensive documentation possible, we thought you may find the below sections a helpful place to get started.

- The [Getting Started](./getting-started/basic-usage.md) section provides onboarding material
- The [Development](./development/setup.md) header is the best place to get started on developing on top of and with methodk8s
- See the [Docs](./index.md) section for a comprehensive rundown of methodk8s capabilities

# About methodk8s

methodk8s provides security operators with a number of data-rich K8s enumeration capabilities to help them gain visibility into their K8s environments. Designed with data-modeling and data-integration needs in mind, methodk8s can be used on its own as an interactive CLI, orchestrated as part of a broader data pipeline, or leveraged from within the Method Platform.

The number of security-relevant K8s resources that methodk8s can enumerate are constantly growing. For the most up to date listing, please see the documentation [here](./index.md)

To learn more about methodk8s, please see the [Documentation site](https://method-security.github.io/methodk8s/) for the most detailed information.

## Quick Start

### Get methodk8s

For the full list of available installation options, please see the [Installation](./getting-started/installation.md) page. For convenience, here are some of the most commonly used options:

- `docker run methodsecurity/methodk8s`
- `docker run ghcr.io/method-security/methodk8s`
- Download the latest binary from the [Github Releases](https://github.com/Method-Security/methodk8s/releases/latest) page
- [Installation documentation](./getting-started/installation.md)

### Authentication
Authentication can be done in 3 ways:
1. By setting the `--path` flag to the path of a `.kube/config` file
2. Setting the `$KUBECONFIG` env variable to the path of a kube config file
3. Configure a Sevice Account (See Service Account documentation for more detail)

### General Usage

```bash
methodk8s <resource> enumerate 
```

#### Examples

```bash
methodk8s pod enumerate --url test-cluster.net
```

```bash
methodk8s node enumerate --context minikube --path ~/kube/.config
```

## Contributing

Interested in contributing to methodk8s? Please see our organization wide [Contribution](https://method-security.github.io/community/contribute/discussions.html) page.

## Want More?

If you're looking for an easy way to tie methodk8s into your broader cybersecurity workflows, or want to leverage some autonomy to improve your overall security posture, you'll love the broader Method Platform.

For more information, visit us [here](https://method.security)

## Community

methodk8sis a Method Security open source project.

Learn more about Method's open source source work by checking out our other projects [here](https://github.com/Method-Security) or our organization wide documentation [here](https://method-security.github.io).

Have an idea for a Tool to contribute? Open a Discussion [here](https://github.com/Method-Security/Method-Security.github.io/discussions).
