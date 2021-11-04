package handler

type GetAllGatewaysCmd struct {
}

func (cmd *GetAllGatewaysCmd) Run() error {

	iotcockpit, _ := GetConfig()
	iotcockpit.Debug = true
	iotcockpit.GetAllGateways()
	return nil
}
