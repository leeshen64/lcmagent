Demeter CPE Agent
===
About
---
[Demeter CPE Agent](#demeter-cpe-agent) is a daemon application for DSL devices to communicate with [Demeter](https://github.com/sercomm-cloudwu/lcmdemeter) server

Features
---
* Establish communications between Demeter server and DSL devices via WebSocket
* Convert commands from Demeter server to [ubus](https://git.openwrt.org/project/ubus.git) on DSL devices

Getting Start
---
1. Install [Go](https://golang.org/) command line environment
2. Build the project by `go build` directly or cross-build for specific target platform by `cpe_agent/Makefile`, e.g. For ARM version 7 platform:
```console
$ cd cpe_agent
$ make arm.v7
```
3. Adjust the configuration file in `conf/cpe_agent.yaml`

| Key Name    | Description                                                     |
| ----------- | --------------------------------------------------------------- |
| host        | Address of your Demeter server                                  |
| port        | TCP port of your Demeter server                                 |
| path        | WebSocket endpoint of your Demeter server                       |
| enableSSL   | Specific the SSL of WebSocket endpoint should be enabled or not |
| folder      | Folder path where the log file will be placed                   |
| rotateCount | Log file auto rotate count                                      |

4. Launch CPE agent
```console
$ ./cpe_agent -c ./conf/cpe_agent.yaml
```
5. Command line arguments

| Argument | Description                                     |
| -------- | ----------------------------------------------- |
| h        | Show help                                       |
| v        | Show version information                        |
| d        | Launch agent application as a background daemon |
| c        | Specific path of configuration file             |