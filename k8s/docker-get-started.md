# Docker 使用入门

上篇文章介绍了容器的基本概念，以及 Docker 是目前最主流的容器平台以及 Docker 的镜像、镜像仓库和容器。这篇文章将介绍 Docker 的基本使用，包括镜像的构建、推送以及容器的运行。

在进入正式的主题之前，你应该先安装 Docker 环境。Docker 分为社区版本(Docker CE)和企业版本(Docker EE)，这里我们安装社区版本就可以了。你可以在官网上找到安装指引 https://docs.docker.com/install/。

*如果你使用的是 Windows 操作系统，需要是 Windows 10 企业版、教育版或者专业版本。如果不是这些版本，你可以通过虚拟机或者 WSL2 来使用 Docker。你最好是在 Linux 或者 Mac 操作系统上使用 Docker。*

## 运行官方的 hello-world 镜像

在安装完 docker 之后，你可以使用官方的 hello-world 镜像来测试 docker 是否正确安装:

```sh
$ sudo docker run hello-world
Unable to find image 'hello-world:latest' locally
latest: Pulling from library/hello-world
1b930d010525: Pull complete
Digest: sha256:41a65640635299bab090f783209c1e3a3f11934cf7756b09cb2f1e02147c6ed8
Status: Downloaded newer image for hello-world:latest

Hello from Docker!
This message shows that your installation appears to be working correctly.

To generate this message, Docker took the following steps:
 1. The Docker client contacted the Docker daemon.
 2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    (amd64)
 3. The Docker daemon created a new container from that image which runs the
    executable that produces the output you are currently reading.
 4. The Docker daemon streamed that output to the Docker client, which sent it
    to your terminal.

...
```

上面的过程展示了运行镜像的命令，以及运行镜像之后的输出。如果你看到这些输出，说明你的 docker 安装是正确的。上面展示的信息包括：

* 我们输入了 `sudo docker run hello-world` 命令，然后期望 hello-world 镜像被运行。`hello-world` 为镜像名称，完整的镜像名称包括名字和 tag，比如 `hello-world:0.0.1`，如果我们指定 tag，那么 Docker 使用 `latest` 做为 tag 名。
* docker 发现本地没有 `hello-world:latest` 镜像，然后从镜像仓库中拉取这个镜像。
* 镜像拉取完成后，`hello-world` 容器开始运行，这个容器中运行的进程在标准输出中打印了一段很长的文本内容。表示镜像运行成功了。

现在我们可以通过 `docker image ls` 来查看本地的镜像列表:

```sh
$ sudo docker image ls
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
hello-world         latest              fce289e99eb9        6 months ago        1.84kB
```

以及 `docker container ls` 来查看当前的容器列表，这个命令默认只会显示正在运行中的容器，加上 `--all` 标签才可以展示包括已退出的容器。

```sh
$ sudo docker container ls --all
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS                      PORTS               NAMES
488d2a203477        hello-world         "/hello"            2 minutes ago      Exited (0) 2 minutes ago                       friendly_babbage
```

这展示了我们主机上有 1 个容器，他是基于镜像 `hello-world` 创建的，并且容器名字是 docker 为我们随机生成的 `friendly_babbage`(我们也可以通过参数自己指定容器名字)。

## 构建镜像

确认 Docker 能够正常运行之后，我们可以开始构建自己的镜像了。

### 编写服务程序

我们用 go 语言写一个简单的 http 服务程序，并且命名为 `main.go`，为每个访问的用户输出服务主机名以及服务程序启动命令:

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
		w.Write([]byte(fmt.Sprintf("服务器的主机名为: %s 服务启动命令为: %v", host, args)))
	}))

	if err := http.ListenAndServe("0.0.0.0:80", nil); err != nil {
		panic(err)
	}
}
```

### 编写 Dockerfile

为了把应用打包成镜像，首先我们需要创建 Dockerfile 文件

```Dockerfile
FROM golang:1.12
WORKDIR /app
ADD main.go .
RUN go build -o app 

CMD ./app 
```

* `FROM golang:1.12` 表示我们需要基于 `golang:1.12` 镜像来构建我们的镜像， Docker 的镜像是一层一层的，在 Dockerfile 中，每条命令都会基于上一层来创建一个新层(极个别命令以及注释除外)，不同的镜像可以共享相同的分层（比如例子中的 golang:1.12 被所有基于它的镜像共享）。
* `WORKDIR /app` 命令将工作目录设置为 `/app`，这条命令过后，它也会创建一个镜像层，并且每个基于此镜像的镜像的工作目录都会是 `/app`，除非再次使用 `WORKDIR` 显示修改。注意：由于 Dockerfile 中的每条指令都是独立执行的，你不能使用后面将要讲到的 `RUN` 命令来设置 （比如：`RUN cd /app`）工作目录
* `ADD main.go .` 也会创建一个新层，并且将当前目录下的 main.go 文件拷贝到镜像中
* `RUN go build -o app ` 在镜像的新层中执行 `go build -o app` 命令，编译我们的程序，并且输出为 app 文件
* `CMD ./app` 也会创建一个新增，但是不会在构建镜像的时候运行 `./app`，它指示 docker 启动此镜像时默认执行的命令。

*注意：不要混淆 CMD 和 RUN，RUN 会在构建镜像的过程中执行，CMD 不会在构建镜像的过程中执行后面的命令，它告诉 Docker 默认的启动命令是什么。*

这里只根据我们的例子介绍了 Dockerfile 的几个常用的指令，完整的 Dockerfile 教程可以参考官方文档：<https://docs.docker.com/engine/reference/builder/>

### docker build

一旦我们编写好了 Dockerfile 和我们的服务程序，就可以使用 `docker build` 命令构建镜像了。我们把 Dockerfile 和 main.go 放在同一个目录中，然后执行 `docker build . -t myapp `。第一个参数表示要用哪个目录来创建镜像，第二个参数 `-t myapp` 表示镜像的名字，我们没有指定 tag，所以 Docker 会默认指定 tag 为 `latest`。

下面展示了构建镜像时 Docker 的输出：

```shell
$ ls
Dockerfile  main.go

$ sudo docker build . -t myapp
Sending build context to Docker daemon  3.072kB
Step 1/5 : FROM golang:1.12
1.12: Pulling from library/golang
a4d8138d0f6b: Pull complete
dbdc36973392: Pull complete
f59d6d019dd5: Pull complete
aaef3e026258: Pull complete
0131e4edf1f3: Pull complete
8013cb24ecbc: Pull complete
f4fcc76edb41: Pull complete
Digest: sha256:7376df6518e9cf46872c5c6284b9787b02e7b3614e45ff77acc9be0d02887ff1
Status: Downloaded newer image for golang:1.12
 ---> f50db16df5da
Step 2/5 : WORKDIR /app
 ---> Running in ca95a898174c
Removing intermediate container ca95a898174c
 ---> 0470eb62d6f5
Step 3/5 : ADD main.go .
 ---> 863941cbb625
Step 4/5 : RUN go build -o app
 ---> Running in cce4a39487c0
Removing intermediate container cce4a39487c0
 ---> cd8cda1acde7
Step 5/5 : CMD ./app
 ---> Running in 0416de1323ca
Removing intermediate container 0416de1323ca
 ---> f64d1428d623
Successfully built f64d1428d623
Successfully tagged myapp:latest
```

有必要先指出的是，Docker 使用的是 CS（客户端/服务器） 架构，在安装完 Docker 后，会启动一个 Docker 守护进程，它相当于服务器。`docker image`、`docker container`、`docker build` 等命令，都是客户端命令，它和服务器交互，并且输出操作结果。

当执行 build 时，docker 客户端将当前目录（由第一个参数指定的）的所有文件发送到服务器，然后服务器开始构建镜像。由于 Docker 会将所有内容都发给服务器，所以应当不要在当前目录放一些不必要的文件。

接下来，服务器开始构建镜像，它一条一条地执行 Dockerfile 中的指令，每条命令都会产生一个新的镜像层。 例如执行 `FROM golang:1.12` 时，由于本地没有 golang:1.12 镜像，所以会先从镜像仓库中拉取镜像，拉取的过程中也是一层一层的处理的，如果本地已经有对应的层了，那么就不会再拉取。再如 `ADD` 命令，它将 main.go 文件拷贝到镜像的工作目录中，这也会产生一个新的层。

镜像构建成功后，可以使用 `docker image ls` 命令查看当前的镜像列表：

```shell
$ sudo docker image ls
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
myapp               latest              7b3f0b9b4107        About 2 minutes ago   781MB
```

使用 `docker image tag myapp <yourname>/myapp` 为镜像指定一个新的名字，这样就可以将镜像推送到官方的镜像仓库了: `docker push <yourname>/myapp`（在推送之前，你需要先在 https://hub.docker.com 创建账号）

## 运行容器

有了镜像，现在我们可以开始运行自己的镜像了:

```shell
$ sudo docker run --name myapp_container -p8080:80 -d myapp
```

此命令基于镜像 myapp 启动一个新容器，这个容器与终端分离 (-d 标志)，这意味着它在后台运行。容器的名字由 `--name` 参数指定为 `myapp_container`（如果没有此参数， Docker 会为我们生成一个随机的名字）。`-p8080:80` 参数将主机的 8080 端口映射到容器的 80 端口。如果你将镜像推送到了镜像仓库，也可以在任何安装了 Docker 的机器上这样启动容器： `docker run --name myapp_container -p8080:80 -d <yourname>/myapp`

可以通过 `http://localhost:8080` 来访问我们的服务： 

```shell
$ curl http://localhost:8080
服务器的主机名为: 853f812a44bb 服务启动命令为: [./app]
```

可以看到我们的服务运行在一个主机名为 853f812a44bb 的主机上，它的启动命令是就是 Dockerfile 中使用 CMD 指定的 ./app。也可以这样启动容器: `sudo docker run --name myapp_container -p8080:80 -d myapp param1 param2`，这样对应的输出就会变成 `[./app param1 param2]`，因为 `docker run` 会将镜像名称后面的所有参数都作为启动命令的额外参数。

通过 `docker container ls` 或者 `docker ps` 都可以看到当前运行中的容器：

```shell
$ sudo docker container ls
CONTAINER ID        IMAGE               COMMAND              CREATED             STATUS              PORTS                  NAMES
853f812a44bb        myapp               "/bin/sh -c ./app"   3 minutes ago       Up 3 minutes        0.0.0.0:8080->80/tcp   myapp_container
```

通过 `docker exec -it myapp_container bash` 可以进入容器内部。－it 选项是下面两个选项的简写：

* -i，确保标准输入流保持开放。 需要在 shell 中输入命令。 
* -t，分配一个伪终端（TTY）。 

通过 ps 命令可以看到当前容器内运行的进程：

```bash
$ docker exec -it myapp_container bash
root@853f812a44bb:/app#
root@853f812a44bb:/app# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.0   4280   736 ?        Ss   01:35   0:00 /bin/sh -c ./app
root         6  0.0  0.0 219544  8608 ?        Sl   01:35   0:00 ./app
root        12  0.6  0.0  18188  3232 pts/0    Ss   01:36   0:00 bash
root        17  0.0  0.0  36636  2776 pts/0    R+   01:37   0:00 ps aux
```

容器中启动了 4 个进程:

* /bin/sh -c ./app： 由于 Dockerfile 中的 CMD 的形式是 `shell form`，所以容器运行时启动的是一个 sh 进程，它用 -c 选项加上参数 ./app 启动 ./app。
* ./app： 这就是我们期望在容器内部运行的进程
* bash： 我们刚才通过 exec 命令创建的进程
* ps aux：和往常一样， ps 命令会输出它自己

当要关闭服务时，可以使用 `docker stop` 命令停止容器的运行， `docker rm` 命令移除容器（移除后才是真正的删除，它会删除容器内的所有文件）：

```sh
$ sudo docker stop myapp_container
myapp_container

$ sudo docker rm myapp_container
myapp_container
```

同一个镜像可以在同一台机器中运行多个不同的容器，它们的文件系统、进程树等都是独立的。容器使用了类似写时复制的技术，初始时它基于镜像创建了一个运行中的容器，它的所有文件和镜像一样，但是当它需要修改某个文件时，就会在容器内部创建这个文件的一份拷贝，而镜像中的文件不受影响。

