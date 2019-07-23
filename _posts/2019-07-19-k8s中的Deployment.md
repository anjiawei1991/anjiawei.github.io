---
layout: post
author: 安佳玮
---

前一篇文章介绍了如何使用 kubernetes 创建 pod。但是在正式环境中，一般不会直接创建 pod 对象，而是使用其它更加高层的资源托管 pod。这就是这篇文章要探索的主题：Deployment。通过 Deployment，可以：

* 水平扩展应用，使得 pod 可以拥有多个副本，并且集群会维持期望的副本数量，比如在 pod 异常退出时重新创建
* 通过设置镜像版本号，自动完成滚动升级
* 在滚动升级过程中暂停升级或者回滚

尽管 Deployment 是使用 ReplicaSet 控制器来达成自己的目的的，但是本文会尽量避免讲述这些概念，将核心焦点放在 Deployment 提供的上层功能上。

# 1. 创建 Deployment

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
              image: user/myapp:0.0.2
              ports:
              - containerPort: 80
                protocol: TCP
```

以上代码展示了一个简单的 Deployment 描述文件，和之前定义的 pod 描述文件一样，主要分为 3 个部分： 

1. 基础信息，apiVersion 定义了所属 API 组(apps)和版本(v1)，kind 指示的资源类型 Deployment
2. metadata，这和 pod 的 metadata 一样，用于定义一些元数据，比如资源的名字
3. spec 表示 Deployment 的详细内容:
    * `replicas` 代表所期望的副本数量
    * `selector` 表示选择器，它通过附加在 pod 上的标签来确定这个 pod 是否隶属于自己，后文会继续介绍这个属性的作用。
    * `template` 为所需要部署的 pod 的模板，这里的内容和上篇文章中的 pod 描述文件类似，Deployment 使用此模板创建 pod。
    
你可能注意到了，这里的 template 比之前的 pod 的描述文件多了一个 labels 字段。这个 labels 表示 pod 的标签。
尽管 pod 和 Deployment 放在同一个描述文件中，但它们在集群中实际上是不同的对象。所以需要一种机制来将它们关联，labels 就是这样的一种机制。Deployment 通过自己的选择器来和 pod 进行匹配。

将描述文件保存为 `myapp-deployment.yaml`，然后使用 `kubectl apply -f myapp-deployment` 命令应用此文件。apply 命令会检查文件中的资源是否存在，如果不存在，就会创建这个资源；如果资源已经存在了，那么就会更新它。

```s
$ kubectl apply -f myapp-deployment.yaml
deployment.apps/myapp-deployment created
```

这里 apply 执行时，由于名字为 myapp-deployment 的对象不存在，所以集群会创建这个 Deployment 对象。通过 `kubectl get deployment` 查看此对象:

```s
$ kubectl get deployment
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
myapp-deployment   0/3     3            0           20s
```

从这个简要的状态可以看到，myapp-deployment 刚创建完成，现在已经准备好的 pod 数量为 0。现在来看一下 pod 列表：

```s
$ kubectl get pods --watch
NAME                                READY   STATUS              RESTARTS   AGE
myapp-deployment-7558d88658-p5slp   0/1     ContainerCreating   0          29s
myapp-deployment-7558d88658-snqnq   0/1     ContainerCreating   0          29s
myapp-deployment-7558d88658-tpt55   0/1     ContainerCreating   0          29s
myapp-deployment-7558d88658-tpt55   1/1     Running             0          32s
myapp-deployment-7558d88658-p5slp   1/1     Running             0          34s
myapp-deployment-7558d88658-snqnq   1/1     Running             0          35s
```

由于要持续监控 pod 的状态变更，所以 get pods 的时候，传递了 --watch 标志。这样 pod 的状态变更时，会即时在屏幕上输出。
从上面的信息来看，有 3 个独立的 pod，一开始看到它们的状态为 ContainerCreating，然后变成了 Running。pod 的名字是集群根据 deployment 和模板信息自动生成的。
那现在再来看一眼 deployment 的状态：

```s
$ kubectl get deployment
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
myapp-deployment   3/3     3            3           45s
```

现在已经有 3 个已经准备好了而且可用的 pod 了。

# 2. 自动拉起 

Deployment 会持续监控管理的 pod 的状态，它通过 selector 来找出和自己匹配的 pod，如果 pod 的数量过多，它就会将多余的部分杀掉。如果 pod 的数量过少，它就会拉起新的 pod 以满足所期望的状态。

现在通过杀掉一个 pod 来模拟 pod 异常退出的情况：

```s
$ kubectl delete pod myapp-deployment-7558d88658-p5slp
pod "myapp-deployment-7558d88658-p5slp" deleted

$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
myapp-deployment-7558d88658-r2nzx   1/1     Running   0          14s
myapp-deployment-7558d88658-snqnq   1/1     Running   0          1m
myapp-deployment-7558d88658-tpt55   1/1     Running   0          1m
```

上面的 delete 命令将名字为 `myapp-deployment-7558d88658-p5slp` 的 pod 杀掉了，使用 `kubectl get pods` 获取当前的 pod 列表时，
pod 的数量仍然为 3，但刚才杀掉的那个 pod 不见了，它被替换成了一个新的 pod: `myapp-deployment-7558d88658-r2nzx`。
这是因为 Deployment 有自动拉起机制，它通过 selector 持续匹配当前的 pod 列表，如果检测到 pod 的数量不足了，就自动创建。

那么在 pod 数量过多的时候，Deployment 是不是也会杀掉多余的 pod 呢？来试试看。通过上篇文章的 myapp-pod.yaml 来创建一个 pod：

```s
$ kubectl apply -f myapp-pod.yaml
pod/myapp created

master $ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
myapp                               1/1     Running   0          15s
myapp-deployment-7558d88658-r2nzx   1/1     Running   0          1m
myapp-deployment-7558d88658-snqnq   1/1     Running   0          2m
myapp-deployment-7558d88658-tpt55   1/1     Running   0          2m
```

创建 pod 之后，pod 列表中包含了 4 个 pod。Deployment 并没有杀掉多余的 pod？

现在来分析一下为什么，这是由于 Deployment 是通过 selector 来匹配 pod 的，但是手动创建的这个 pod 并没有被 Deployment 匹配到。因此它不会被视为 Deployment 的管理范畴。

**Deployment 总是通过 selector 来确定属于它的 pod，这是 Kubernetes 应用得非常广泛的一种机制：对象之间本身并没有关联，而是通过标签和选择器关联到了一起**

通过 --show-labels 参数可以看到 pod 的标签：

```s
$ kubectl get pods --show-labels
NAME                                READY   STATUS    RESTARTS   AGE   LABELS
myapp                               1/1     Running   0          58s   <none>
myapp-deployment-7558d88658-r2nzx   1/1     Running   0          2m    app=myapp,pod-template-hash=7558d88658
myapp-deployment-7558d88658-snqnq   1/1     Running   0          3m    app=myapp,pod-template-hash=7558d88658
myapp-deployment-7558d88658-tpt55   1/1     Running   0          3m    app=myapp,pod-template-hash=7558d88658
```

手动创建的 pod 并没有标签，而 Deployment 自动创建的 pod 都包含了两个标签，其中 1 个为 app=myapp，这是描述文件中指定的。
另外还有一个多余的 pod-template-hash，它表示 Deployment 中 pod 模板的 hash。

通过 labels 可以更新 pod 的标签。现在给手动创建的 pod 附上这两个标签，再看看 pod 列表：

```s
$ kubectl label pods myapp app=myapp pod-template-hash=7558d88658
pod/myapp labeled

$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
myapp-deployment-7558d88658-r2nzx   1/1     Running   0          2m
myapp-deployment-7558d88658-snqnq   1/1     Running   0          3m
myapp-deployment-7558d88658-tpt55   1/1     Running   0          3m
```

手动创建的那个名字为 myapp 的 pod 消失了，它是被 Deployment 自动删除的。这印证了刚才说过的：**Deployment 总是通过 selector 来确定属于它的 pod**

delete 子命令也可以指定一个描述文件，删除描述文件中的对象。在删除 Deployment 的同时，它所管理的 pod 也会被自动删除掉：

```s
$ kubectl delete -f myapp-deployment.yaml
deployment.apps "myapp-deployment" deleted

$ kubectl get pods
No resources found.
```

# 3. 水平扩容

基于 Deployment，可以很方便的实现水平扩容和缩容。首先再次使用 apply 创建 Deployment:

```s
$ kubectl apply -f myapp-deployment.yaml
$ kubectl get pods
myapp-deployment-7558d88658-tjfwn   1/1     Running             0          49s
myapp-deployment-7558d88658-f7stc   1/1     Running             0          51s
myapp-deployment-7558d88658-gn9k6   1/1     Running             0          52s
```

假设应用现在在生成环境下跑了一段时间了，由于过于火爆，需要进行扩容。经过讨论，决定将应用数量水平扩容到 4 个。
水平扩容，只需做非常少量的工作，将 replicas 字段修改为 4，其它内容保持不变：

```s
$ sed -i 's/replicas: 3/replicas: 4/' myapp-deployment.yaml
```

接着，再次使用 apply 命令即可完成扩容：

```s
$ kubectl apply -f myapp-deployment.yaml
deployment.apps/myapp-deployment configured

$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
myapp-deployment-7558d88658-f7stc   1/1     Running   0          2m20s
myapp-deployment-7558d88658-gn9k6   1/1     Running   0          2m20s
myapp-deployment-7558d88658-rbbq7   1/1     Running   0          6s
myapp-deployment-7558d88658-tjfwn   1/1     Running   0          2m20s
```

可以看到，除了之前创建的 3 个 pod 之外，又多了 1 个新的 pod。

由于运营策略的调整，公司要将用户引流到其它产品。为了避免资源浪费，所以现在要进行缩容，这和扩容一样简单。再次将 replicas 修改为 3，然后执行 apply 命令：


```s
$ sed -i 's/replicas: 4/replicas: 3/' myapp-deployment.yaml

$ kubectl apply -f myapp-deployment.yaml
deployment.apps/myapp-deployment configured

$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
myapp-deployment-7558d88658-f7stc   1/1     Running   0          2m58s
myapp-deployment-7558d88658-gn9k6   1/1     Running   0          2m58s
myapp-deployment-7558d88658-tjfwn   1/1     Running   0          2m58s
```

一小段时间后，pod 的数量又变回了 3，其中的一个 pod 被删除了。

# 4. 升级和回滚

上面的内容讨论过了如果创建 Deployment，以及介绍了 Deployment 能够自动调整 pod 的数量达到期望状态。接下来将会介绍如何使用 Deployment 进行升级和回滚操作。

## 4.1 构建新版本镜像

先更新服务程序，同时让它打印版本号 0.0.2:

```go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		args := os.Args
		host, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		w.Write([]byte(fmt.Sprintf("服务器的主机名为: %s 服务启动命令为: %v 当前版本号为: 0.0.2", host, args)))
	}))

	if err := http.ListenAndServe("0.0.0.0:80", nil); err != nil {
		panic(err)
	}
}
```

然后使用 `docker build . -t user/myapp:0.0.2` 构建新版本的镜像，并且通过 `docker push user/myapp:0.0.2` 将镜像推送到镜像仓库中。

## 4.2 升级应用

接下来，修改描述文件，将镜像版本号修改为 0.0.2:

```s
$ sed -i 's/image: user\/myapp:0.0.1/image: user\/myapp:0.0.2/g' myapp-deployment.yaml
```

还是使用 apply 命令，应用新版本的描述文件：

```s
$ kubectl apply -f myapp-deployment.yaml
deployment.apps/myapp-deployment configured
```

查看 pod:

```s
$ kubectl get pods
NAME                                READY   STATUS              RESTARTS   AGE
myapp-deployment-674c68985-55j7m    0/1     ContainerCreating   0          6s
myapp-deployment-7558d88658-f7stc   1/1     Running             0          4m13s
myapp-deployment-7558d88658-gn9k6   1/1     Running             0          4m13s
myapp-deployment-7558d88658-tjfwn   1/1     Running             0          4m13s

$ kubectl get pods
NAME                               READY   STATUS    RESTARTS   AGE
myapp-deployment-674c68985-55j7m   1/1     Running   0          78s
myapp-deployment-674c68985-l79xv   1/1     Running   0          71s
myapp-deployment-674c68985-mtdws   1/1     Running   0          69s
```

一段时间后，旧版本的 pod 全部被删除了，取而代之的是 3 个新的 pod。查看其中任何一个 pod 的内容，发现它的镜像已经更新为 user/myapp:0.0.2


```s
$ kubectl get pods myapp-deployment-674c68985-55j7m  -o go-template="{{range .spec.containers}} {{.image}} {{end}}"
user/myapp:0.0.1
```

这里的 `-o go-template=...` 使用 [go 模板](https://golang.google.cn/pkg/text/template/) 输出指定的内容。
如果要查看镜像，就使用 `{{range .spec.containers}} {{.image}} {{end}}`，它遍历 .spec 域下的所有容器，并且输出 .image 字段的值。

此时，查看升级状态：

```s
$ kubectl rollout status deployment myapp-deployment
deployment "myapp-deployment" successfully rolled out
```

表明已经成功升级完成。rollout 子命令可以用来查看资源的更新状态、更新历史等。下面查看历史版本：

```s
$ kubectl rollout history -f myapp-deployment.yaml
deployment.apps/myapp-deployment
REVISION  CHANGE-CAUSE
1         <none>
2         <none>
```

输出结果显示 Deployment 包含了两个历史版本，1 表示的是第一次创建 Deployment 时的版本，2 表示的是刚才的升级的版本。 CHANGE-CAUSE 这一列是这个版本对应的命令，只有在命令使用了 --record 参数时才会被记录，比如：

```s
$ kubectl apply -f myapp-deployment.yaml --record
deployment.apps/myapp-deployment created

$ kubectl rollout history -f myapp-deployment.yaml
deployment.apps/myapp-deployment
REVISION  CHANGE-CAUSE
1         kubectl apply --filename=myapp-deployment.yaml --record=true
```

## 4.3 升级策略

Deployment 目前支持 2 种不同的升级策略，分别为 RollingUpdate 和 Recreate，默认的策略为 RollingUpdate。

* RollingUpdate 表示的是滚动更新，它通过逐步启动新版本、关闭旧版本的方式来更新应用。由于更新过程中同时运行了新版本和旧版本，这要求应用可以同时兼容新旧版本。另外，使用这种方式可以实现不停机更新，在更新过程中一直有 pod 在提供服务。
* Recreate 表示的是重新创建。它先将旧版本的 pod 删掉，然后再创建新版本。这种更新模式不需要应用同时兼容新旧版本，但是在更新过程中会出现服务不可用的状态。

上面演示的是使用了默认的升级策略 RollingUpdate。如果是下面的描述文件，升级策略将会是 Recreate：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: myapp-deployment
spec:
    replicas: 3
    selector:
        matchLabels:
            app: myapp
    strategy:
        type: Recreate
    template:
        metadata:
            labels:
                app: myapp
        spec:
            containers:
            - name: myapp
              image: user/myapp:0.0.2
              ports:
              - containerPort: 80
                protocol: TCP
```

## 4.4 回滚应用

Deployment 的回滚也是非常简单的，当新版本出现故障，可能需要回滚。这只需要一条命令: `rollout undo`。执行它然后查看 pod 列表和最新的版本号：

```s
$ kubectl rollout undo deployment myapp-deployment
deployment.extensions/myapp-deployment rolled back

$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
myapp-deployment-7558d88658-2gswh   1/1     Running   0          25s
myapp-deployment-7558d88658-8chtx   1/1     Running   0          23s
myapp-deployment-7558d88658-qrd6p   1/1     Running   0          21s

$ kubectl get pods myapp-deployment-7558d88658-2gswh  -o go-template="{{range .spec.containers}} {{.image}} {{end}}"
user/myapp:0.0.1
```

undo 命令默认会回滚到最近的上一个版本，也可以使用 `--to-revision=<some-revision>` 来回滚到任意一个版本。上面的信息展示了版本的确回滚到 0.0.1 了，再看一下 Deployment 的历史记录：

```s
$ kubectl rollout history deployment myapp-deployment
deployment.extensions/myapp-deployment
REVISION  CHANGE-CAUSE
2         <none>
3         <none>
```

现在，仍然只看到了 2 个版本号。这是因为 Deployment 知道版本 3 就是通过回滚的版本 1，所以它会将版本 1 移除掉。Deployment 中默认记录最近 10 个版本的历史，通过回滚更新的版本只算 1 个。


## 4.5 灰度发布

为了避免新版本直接上线可能导致的异常，实际更新时往往需要灰度发布。Deployment 没有直接提供对灰度发布功能的支持，但是可以通过在更新过程中暂停和继续来实现类似的效果。灰度发布是在升级策略为 RollingUpdate 的情况下讨论的。

在继续这个主题之前，需要了解的是 pod 更新速度可能非常快，在还没反应过来之前，可能就已经更新完成了。
所以需要设置 minReadySeconds，这个属性指定新创建的 pod 至少成功运行多长时间之后，才将其视为可用。滚动更新也会通过确认新 pod 可用时才会继续进行下一步操作。

现在将 minReadySeconds 属性设置为 30，表示 pod 至少成功运行 30s 才确定为可用的：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: myapp-deployment
spec:
    replicas: 3
    selector:
        matchLabels:
            app: myapp
    minReadySeconds: 30
    strategy:
        type: RollingUpdate 
    template:
        metadata:
            labels:
                app: myapp
        spec:
            containers:
            - name: myapp
              image: user/myapp:0.0.2
              ports:
              - containerPort: 80
                protocol: TCP
```

同样的，使用 apply 子命令进行升级：

```s
$ kubectl apply -f myapp-deployment.yaml
deployment.apps/myapp-deployment configured

$ kubectl get pods 
NAME                                READY   STATUS    RESTARTS   AGE
myapp-deployment-674c68985-bxfkk    1/1     Running   0          5s
myapp-deployment-7558d88658-2gswh   1/1     Running   0          3m15s
myapp-deployment-7558d88658-8chtx   1/1     Running   0          3m13s
myapp-deployment-7558d88658-qrd6p   1/1     Running   0          3m11s
```

观察到在升级过程中出现了 4 个 pod。滚动更新会短暂的出现多余的 pod，因为它要创建新 pod，等待确认新 pod 可用后再删除旧 pod。
能够多出的 pod 数量可以通过 `maxSurge` 属性设置，本文不打算介绍这个属性，有兴趣的读者可以参考[官方文档](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.15/#deployment-v1-apps)


通过 `rollout status` 来查看更新状态，在未更新完成前，这个命令会输出当前状态并且持续等待。等到 2 个新版本的 pod 加入时，使用 `CTRL+C` 退出此命令，然后准备暂停更新：

```s
$ kubectl rollout status deployment myapp-deployment
Waiting for deployment "myapp-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "myapp-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "myapp-deployment" rollout to finish: 1 out of 3 new replicas have been updated...
Waiting for deployment "myapp-deployment" rollout to finish: 2 out of 3 new replicas have been updated...
Waiting for deployment "myapp-deployment" rollout to finish: 2 out of 3 new replicas have been updated...
^C
```

通过 `rollout pause` 命令可以暂停更新：

```s
$ kubectl rollout pause deployment myapp-deployment
deployment.extensions/myapp-deployment paused

$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
myapp-deployment-674c68985-8qzdc    1/1     Running   0          67s
myapp-deployment-674c68985-bxfkk    1/1     Running   0          2m8s
myapp-deployment-7558d88658-2gswh   1/1     Running   0          5m18s
myapp-deployment-7558d88658-8chtx   1/1     Running   0          5m16s
```

暂停更新后，发现现在 pod 的数量为 4，它包括 2 个新版本 pod 和 2 个旧版本 pod。在运行一段时间之后，发现新版本没有问题，然后通过 resume 子命令来继续更新剩下的 pod：

```s
$ kubectl rollout resume deployment myapp-deployment
deployment.extensions/myapp-deployment resumed

$ kubectl rollout status deployment myapp-deployment
Waiting for deployment "myapp-deployment" rollout to finish: 1 old replicas are pending termination...
deployment "myapp-deployment" successfully rolled out
```

至此，所有的 pod 已经更新完成。


# 5. 总结

本文介绍了 Deployment 的使用，它提供了很多高层次的功能，比如水平扩容、缩容，应用的更新、回滚等。另外，通过暂停和继续更新的功能，可以间接的实现灰度发布的功能。