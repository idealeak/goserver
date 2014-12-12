// transhandler
package transact

type TransHandler interface {
	OnExcute(n *TransNode, ud interface{}) TransExeResult
	OnCommit(n *TransNode) TransExeResult
	OnRollBack(n *TransNode) TransExeResult
	OnChildTransRep(n *TransNode, hChild TransNodeID, retCode int, ud interface{}) TransExeResult
}

type OnExecuteWrapper func(n *TransNode, ud interface{}) TransExeResult
type OnCommitWrapper func(n *TransNode) TransExeResult
type OnRollBackWrapper func(n *TransNode) TransExeResult
type OnChildRespWrapper func(n *TransNode, hChild TransNodeID, retCode int, ud interface{}) TransExeResult

type TransHanderWrapper struct {
	OnExecuteWrapper
	OnCommitWrapper
	OnRollBackWrapper
	OnChildRespWrapper
}

func (wrapper *TransHanderWrapper) OnExcute(n *TransNode, ud interface{}) TransExeResult {
	if wrapper.OnExecuteWrapper != nil {
		return wrapper.OnExecuteWrapper(n, ud)
	}
	return TransExeResult_Success
}

func (wrapper *TransHanderWrapper) OnCommit(n *TransNode) TransExeResult {
	if wrapper.OnCommitWrapper != nil {
		return wrapper.OnCommitWrapper(n)
	}
	return TransExeResult_Success
}

func (wrapper *TransHanderWrapper) OnRollBack(n *TransNode) TransExeResult {
	if wrapper.OnRollBackWrapper != nil {
		return wrapper.OnRollBackWrapper(n)
	}
	return TransExeResult_Success
}

func (wrapper *TransHanderWrapper) OnChildTransRep(n *TransNode, hChild TransNodeID, retCode int, ud interface{}) TransExeResult {
	if wrapper.OnChildRespWrapper != nil {
		return wrapper.OnChildRespWrapper(n, hChild, retCode, ud)
	}
	return TransExeResult_Success
}
