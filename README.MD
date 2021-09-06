[![CircleCI](https://circleci.com/gh/disaster37/kubetool/tree/v1.18.svg?style=svg)](https://circleci.com/gh/disaster37/kubetool/tree/v1.18)
[![Go Report Card](https://goreportcard.com/badge/github.com/disaster37/kubetool)](https://goreportcard.com/report/github.com/disaster37/kubetool)
[![GoDoc](https://godoc.org/github.com/disaster37/kubetool?status.svg)](http://godoc.org/github.com/disaster37/kubetool)
[![codecov](https://codecov.io/gh/disaster37/kubetool/branch/v1.18/graph/badge.svg)](https://codecov.io/gh/disaster37/kubetool/branch/v1.18)

# Kubetool

Extra tools to manage kubernetes.
You can use it to handle patch management on kubernetes nodes. It's attempt to mainly resolve this matter.

## Contribute

You PR are always welcome. Please use the righ branch to do PR:
 - v1.20 for Kubernetes 1.20.X
Don't forget to add test if you add some functionalities.

To build, you can use the following command line:

```sh
make build
```

To lauch golang test, you can use the folowing command line:

```sh
make test
```

## CLI

### Global options

The following parameters are available for all commands line :

- **--kubeconfig**: The kube config file to use. You can also use environment variable `KUBECONFIG`. Default to `$HOME/.kube/config`.
- **--debug**: Enable the debug mode
- **--help**: Display help for the current command

You can set also this parameters on yaml file (one or all) and use the parameters `--config` with the path of your Yaml file.

```yaml
---
kubeconfig: $HOME/.kube/config
```

### List worker nodes

It permit to list all workers nodes. The goal is to loop over to put node on downtime to patch them one by one.
All nodes without labels `master=true` are considered as worker nodes.
It return the list of worker nodes separated by coma.

Sample of command:

```bash
kubetool --kubeconfig "C:\Users\user\.kube\config" list-worker-nodes
```

### List master nodes

It permit to list all master nodes. The goal is to loop over them to put on downtime and then patch one by one.
All node with label `master=true` are considered as master nodes.
It return the list of master nodes separated by coma.

Sample of command:

```bash
kubetool --kubeconfig "C:\Users\user\.kube\config" list-master-nodes
```

### Put node on downtime

It permit to put node on downtime. The goal is to safety patch it and so stop pods before.
It perform the following actions:

- Cordon the node (the node become not schedulable)
- Loop over pod hosted on it, to find namespaces associated to them.
- For each namespace, it will look if configmap called `patchmanagement` with key `pre-job` exist.
  If exist, it will lauch job with the contend oh the key `pre-job` as shell script.
- Drain the node

If you need to run extra actions before stop pods hosted on node, you can add configmap `patchmanagement` on application namespace with the key `pre-script`. If you need expose somes secrets as environment variable to use them on script, you can add the key `secrets` with the list of secret to inject on job.
For exemple, before put on downtime node that hosted elasticsearch statefullset. You should put shard allocation on primary and stop services like ILM, SLM, watcher.

You need to set following parameter:

- **--node-name**: The node name to put on downtime

It return the following code:

- 0: All work fine
- 1: Somethink wrong appear, but the node is uncordonned (shedulable). You can't patch it but you can loop on next node.
- 2: Somethink wrong appear, but the node is cordonned (not schedulable). It's good idea to stop here.

Samble of command:

```bash
kubetool --kubeconfig "C:\Users\user\.kube\config" set-downtime --node-name node-01
```

### Put node online
It permit to put node online after successfully patch it and reboot it.
It perform the following actions:

- Uncordon the node (the node become schedulable)
- Wait 60s to start pods on it if needed
- Loop over pod hosted on it, to find namespaces associated to them.
- For each namespace, it will look if configmap called `patchmanagement` with key `post-job` exist.
  If exist, it will lauch job with the contend oh the key `post-job` as shell script.

If you need to run extra actions after patch it, you can add configmap `patchmanagement` on application namespace with the key `post-script`. If you need expose somes secrets as environment variable to use them on script, you can add the key `secrets` with the list of secret to inject on job.
For exemple, after patch node that hosted elasticsearch statefullset. You should put shard allocation on all and start services like ILM, SLM, watcher.

You need to set following parameter:

- **--node-name**: The node name to put on downtime

It return the following code:

- 0: All work fine
- 1: Somethink wrong appear, but the node is uncordonned (shedulable). You can loop on next node.
- 2: Somethink wrong appear, but the node is cordonned (not schedulable). It's good idea to stop here.

Samble of command:

```bash
kubetool --kubeconfig "C:\Users\user\.kube\config" unset-downtime --node-name node-01
```

### Run patch management pre job

It permit to lauch pre job for patchmanagement on given namespace.

Sample of command:

```bash
kubetool --kubeconfig "C:\Users\user\.kube\config" run-pre-job --namespace test
```

### Run patch management post job

It permit to lauch post job for patchmanagement on given namespace.

Sample of command:

```bash
kubetool --kubeconfig "C:\Users\user\.kube\config" run-post-job --namespace test
```

### Clean evicted pods

It permit to clean all pods that failed because of eviected.

Sample of command:

```bash
kubetool --kubeconfig "C:\Users\user\.kube\config" clean-evicted-pods
```