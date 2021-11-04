package handler

type SendJsonCmd struct {
	In string `required`
}

func (cmd *SendJsonCmd) Run() error {

	return nil
}
