package basic

type termReqCommand struct {
	c *Object
}

func (trc *termReqCommand) Done(o *Object) error {
	if o == nil {
		return nil
	}

	//  When shutting down we can ignore termination requests from owned
	//  objects. The termination request was already sent to the object.
	if o.terminating {
		return nil
	}

	//  If I/O object is well and alive let's ask it to terminate.
	if o.childs.IsExist(trc.c.Id) {
		o.termAcks++
		//  Note that this object is the root of the (partial shutdown) thus, its
		//  value of linger is used, rather than the value stored by the children.
		SendTerm(trc.c)
		//	Remove child
		o.childs.Delete(trc.c.Id)
	}

	return nil
}

func SendTermReq(p *Object, c *Object) bool {
	return p.SendCommand(&termReqCommand{c: c}, false)
}
