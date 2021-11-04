package main

import (
	"github.com/alecthomas/kong"
	handler "github.com/uneidel/sapleonardo/handler"
)

var CLI struct {
	Helper struct {
		FlattenJson handler.FlattenFileCmd `cmd help:"iron your json"`
	} `cmd`
	Sap struct {
		CreateDeviceTemplate handler.Createcmd       `cmd help:"Create a DeviceTemplate"`
		CreateDevice         handler.CreateDeviceCmd `cmd help:"Create a Device"`
	} `cmd`
	Iotcockpit struct {
		GetDevices     handler.GetDevicesCmd     `cmd help:"Get All Devices"`
		GetAllGateways handler.GetAllGatewaysCmd `cmd help:"Get All Gateways"`
	} `cmd`
	Fakelogger struct {
		Send handler.SendJsonCmd `cmd help:"Send via mqttsn/mqtt "`
	} `cmd`
}

func main() {

	ctx := kong.Parse(&CLI,
		kong.Name("sapleonardo"),
		kong.Description("sample leonardo tooling"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: false,
		}))
	ctx.Run()
}
