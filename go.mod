module github.com/touchardv/myhome-presence

go 1.18

replace github.com/JuulLabs-OSS/cbgo => github.com/gkuchta/cbgo v0.0.3-0.20210309070341-a5fcee8c38af

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4
	golang.org/x/sys v0.0.0-20211123173158-ef496fb156ab // indirect
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/JuulLabs-OSS/cbgo v0.0.2
	github.com/muka/go-bluetooth v0.0.0-20220819143208-8b1989180f4c
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/godbus/dbus/v5 v5.0.3 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
