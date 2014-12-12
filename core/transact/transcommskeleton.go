// transcommskeleton
package transact

var txSkeletons = make(map[string]TransactCommSkeleton)

type TransactCommSkeleton interface {
	SendTransResult(parent, me *TransNodeParam, tr *TransResult) bool
	SendTransStart(parent, me *TransNodeParam, ud interface{}) bool
	SendCmdToTransNode(tnp *TransNodeParam, cmd TransCmd) bool
	GetSkeletonID() int
	GetAreaID() int
}

func RegisteTxCommSkeleton(name string, tcs TransactCommSkeleton) {
	if _, exist := txSkeletons[name]; exist {
		panic("repeate registe TxCommSkeleton:" + name)
	}
	txSkeletons[name] = tcs
}

func GetTxCommSkeleton(name string) TransactCommSkeleton {
	if t, exist := txSkeletons[name]; exist {
		return t
	}
	return nil
}
