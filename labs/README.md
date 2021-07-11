# Distributed Systems Labs

## 0. Preparation

I finished these labs in Visual Studio Code, which has a wonderful [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go) provided by Google. However if you use VS Code as you editor to do these labs you may find some troubles installing necessary analysing tools such as `gopkgs`. Probably you may get error information like "tcp dial up time out" in mainland China due to a well-known reason, and I solved it in this way:
1. set a proxy for `go get`
```
$ go env -w GO111MODULE=on
$ go env -w GOPROXY=https://goproxy.io,direct
```

2. click `install all` button to install go tools in VS Code and this time you  should get message `SUCCEED`
3. IMPORTANT: unset `GO111MODULE`, otherwise you may get trouble running the labs as I did(errors like "could not import 'fmt'...")
```
$ go env -w GO111MODULE=off
```

Also I found this [documentation](https://golang.org/doc/code) useful to understand how a Go programme is built and run

## 1. MapReduce