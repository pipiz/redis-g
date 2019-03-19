package command

type Command struct {
	Name string
	Args [][]byte
}

func (command Command) GetArgs() []interface{} {
	args := make([]interface{}, len(command.Args))
	for i, arg := range command.Args {
		args[i] = arg
	}
	return args
}
