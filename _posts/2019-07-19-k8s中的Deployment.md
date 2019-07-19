---
layout: post
author: 安佳玮
---

前一篇文章介绍了如何使用 kubernetes 创建 pod。但是在正式环境中，一般不会直接创建 pod 对象，而是使用其它更加高层的资源托管 pod。这就是这篇文章要探索的主题：Deployment。通过 Deployment，可以：

* 水平扩展应用，使得 pod 可以拥有多个副本，并且集群会维持期望的副本数量，比如 pod 异常退出时重新创建
* 通过设置镜像版本号，自动完成滚动升级
* 在滚动升级过程中暂停升级或者回滚

尽管 Deployment 是使用 ReplicaSet 控制器来达成自己的目的的，但是本文会尽量避免讲述这些概念，将核心焦点放在 Deployment 提供的上层功能上。

# 创建 Deployment

和创建 pod 一样，先提供 Deployment 的描述文件:

```yml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: myapp-deployment
spec:
    replicas: 3
    selector:
        matchLabels:
            app: myapp
    template:
        metadata:
            labels:
                app: myapp
        spec:
            containers:
            - name: myapp
              image: anjiawei/myapp:0.0.1
              ports:
              - containerPort: 80
                protocol: TCP
```

以上代码展示了一个简单的 Deployment 描述文件，和之前定义的 pod 描述文件一样，主要分为 3 个部分： 

1. 基础信息，apiVersion 定义了所属 API 组(apps)和版本(v1)，kind 指示的资源类型 Deployment
2. metadata，这和 pod 的 metadata 一样，用于定义一些元数据，比如资源的名字
3. spec 表示 Deployment 的详细内容:
    * `replicas` 代表所期望的副本数量
    * `selector` 用于描述选择器，选择器通过附加在 pod 上的标签来确定这个 pod 是否隶属于自己。
    * `template` 为所需要部署的 pod 的模板，这里的内容和上篇文章中的 pod 描述文件类似，Deployment 使用此模板创建 pod。
    
你可能注意到了，这里的 template 比之前的 pod 的描述文件多了一个 labels 字段。这个 labels 表示 pod 的标签。Deployment 通过自己的选择器来和 pod 进行匹配。

将描述文件保存为 `myapp-deployment.yaml`，然后使用 `kubectl apply -f myapp-deployment` 命令应用此文件。apply 命令会检查文件中的资源是否存在，如果不存在，就会创建这个资源。

```sh
$ kubectl apply -f myapp-deployment.yaml
deployment.apps/myapp-deployment created
```

<!-- master $ kubectl get deployment
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
myapp-deployment   0/3     3            0           20s
master $ kubectl get pods --watch
NAME                                READY   STATUS              RESTARTS   AGE
myapp-deployment-7558d88658-cfsxc   0/1     ContainerCreating   0          29s
myapp-deployment-7558d88658-clsfg   0/1     ContainerCreating   0          29s
myapp-deployment-7558d88658-m2sfz   0/1     ContainerCreating   0          29s
myapp-deployment-7558d88658-m2sfz   1/1     Running             0          32s
myapp-deployment-7558d88658-cfsxc   1/1     Running             0          34s
myapp-deployment-7558d88658-clsfg   1/1     Running             0          35s
^Cmaster $ kubectl get deployment
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
myapp-deployment   3/3     3            3           45s
master $ vim myapp-deployment.yaml
master $ -->

# 使用 Deployment 升级应用

# 回滚
