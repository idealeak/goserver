package basic

//  Object to process the command.

type Command interface {
	Done(*Object) error
}

type CommandWrapper func(*Object) error

func (cw CommandWrapper) Done(o *Object) error {
	return cw(o)
}
