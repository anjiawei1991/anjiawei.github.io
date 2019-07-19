---
layout: post
author: 安佳玮
---

前一篇文章介绍了如何使用 kubernetes 创建 pod。但是在正式环境中，一般不会直接创建 pod 对象，而是使用其它更加高层的资源托管 pod。

这就是这篇文章要探索的东西：Deployment。Deployment 通过 ReplicaSets 来水平扩展应用，以及提供了一种声明式的升级功能：用户声明所期望的状态， kubernetes 集群自动完成升级或回滚。通过 Deployment，可以：

* 水平扩展应用，使得 pod 可以拥有多个副本，并且集群会维持期望的副本数量，比如 pod 异常退出时重新创建
* 通过设置镜像版本号，自动完成滚动升级
* 在滚动升级过程中暂停升级或者回滚

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
              image: anjiawei/myapp:0.0.2
              ports:
              - containerPort: 80
                protocol: TCP
```

以上代码展示了一个简单的 Deployment 描述文件，和之前定义的 pod 描述文件一样，主要分为 3 个部分： 

1. 基础信息，apiVersion 定义了所属 API 组(apps)和版本(v1)，kind 指示将要创建的资源类型为 Deployment
2. metadata，这和 pod 的 metadata 一样，用于定义一些元数据，比如
3. spec 表示 Deployment 的详细内容:
    * `replicas: 3` 表示所期望的副本数量为 3
    * selector

# 使用 Deployment 升级应用

# 回滚
