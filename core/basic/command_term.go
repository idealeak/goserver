package basic

var termCmd = &termCommand{}

type termCommand struct {
}

func (tc *termCommand) Done(o *Object) error {
	if o == nil {
		return nil
	}

	//  Double termination should never happen.
	o.processTerm()

	return nil
}

func SendTerm(o *Object) bool {
	return o.SendCommand(termCmd, false)
}
