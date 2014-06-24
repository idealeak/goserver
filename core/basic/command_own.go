package basic

type ownCommand struct {
	c *Object
}

func (oc *ownCommand) Done(o *Object) error {

	defer o.ProcessSeqnum()

	//  If the object is already being shut down, new owned objects are
	//  immediately asked to terminate. Note that linger is set to zero.
	if o.terminating {
		o.termAcks++
		SendTerm(o, oc.c)
		return nil
	}

	//  Store the reference to the owned object.
	if o.childs == nil {
		o.childs = make(map[int]*Object)
	}
	o.childs[oc.c.Id] = oc.c

	return nil
}

func SendOwn(p *Object, c *Object) bool {
	return p.SendCommand(c, &ownCommand{c: c}, true)
}
