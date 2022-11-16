# nerdctl-log-example

模拟 `nerdctl run -d --name runcdev1 q946666800/runcdev` 的日志收集过程。

运行 `./shim-example`（使用二进制文件运行）,
首先启动驱动程序，返回相关pio，
```go
pio, err := driveIO()
if err != nil {
	log.Fatal(err)
}
```

配置应用程序，将stdio传递给上面的pio
```go
// start app
cmd := exec.Command("./app-example")
cmd.Stdout = pio.out.w
cmd.Stderr = pio.err.w
```

启动应用程序，日志驱动将应用程序日志，以 `json` 的形式，写入到 `app.log` 文件中。

nerdctl中日志处理方式大致如此。