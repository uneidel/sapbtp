package handler

type GetDevicesCmd struct {
}

func (cmd *GetDevicesCmd) Run() error {

	iotcockpit, _ := GetConfig()
	iotcockpit.GetAllDevices()
	return nil
}
