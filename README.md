# sre-kuby

## Introduction

This go application provides simple and easy way to create kubectl configurations for different servers with google as the oidc (OpenId Connect) provider.

## Prerequisites

* [install go](https://golang.org/doc/install) version 1.13 or newer
* Place the config_auth (replace the client_id, client_secret and the \@example.com domain with your once) file in your .kube folder

```bash
mkdir -p ~/.kube
mv config_auth ~/.kube
```

## Usage

To install the go binary just execute the command below.
If you encounter any problems or errors with this step check the Q&A section.

```bash
export GO111MODULE=on
go get github.com/fr123k/kuby
```

The kuby binary is created in the bin sub-dir of your GOPATH (default is ${HOME}/go/bin).
You can add it to your path or change the path kuby in the commands.
The following command will print the application help manifest.

```bash
kuby --help
```

### One k8s config to rule them all

That is the default list of servers it will try to create the kubectl configuration for:

* dev:k8s-api-dev.example.de
* staging:k8s-api-staging.example.de
* phdp:k8s-api-prod.example.de

To create the kubectl configurations for all know k8s clusters just run the following command.

```bash
kuby
```

The following will happen a web browser will be open and ask for your google account credentials.

After an successfully login the browser window will show a message like.

```text
Authentication completed. It's safe to close this window now ;-)
```

So close the window and follow the instruction on the terminal.

```text
Do you want to overwrite the /xxxxx/xxxxxx/.kube/config. (y/n):
```

After the process is finish just execute kubectl

```bash
kubectl get pods
```

to check if the generated configuration is valid.

### One k8s config per cluster

With the -s or --servers you can specify the k8s api server to create the kubectl configuration for.

```bash
kuby -s dev:k8s-api-dev.example.de
```

You can also create configuration for multiple servers like this.

```bash
kuby -s dev:k8s-api-dev.example.de -s staging:k8s-api-staging.example.de
```

## Q&A

1. go get with private github repositories

   If you see an error like this

   ```text
   fatal: could not read Username for 'https://github.com': terminal prompts disabled
   ```

   The you can enable git ssh for go get with the following command.

   ```bash
   git config --global url."git@github.com:".insteadOf "https://github.com/"
   ```

   For that to work you had to finish the [Github ssh setup](https://help.github.com/en/articles/connecting-to-github-with-ssh)

2. If the `go get github.com/fr123k/kuby` command fails you can download the latest binary from the [`kuby`](https://github.com/fr123k/kuby/releases) GitHub project releases.

   You need to place the binary in the directory `GOPATH/bin` and make the binary executable.
