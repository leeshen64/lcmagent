module sercomm.com/demeter/cpe_agent

go 1.14

replace sercomm.com/demeter/commons => ../commons

require (
	github.com/gorilla/websocket v1.4.2
	github.com/micro/go-micro v1.18.0 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/sevlyar/go-daemon v0.1.5
	go.uber.org/zap v1.16.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	sercomm.com/demeter/commons v0.0.0-00010101000000-000000000000
)
