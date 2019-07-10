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
