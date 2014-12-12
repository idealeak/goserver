package transact

const (
	TransactCommitPolicy_SelfDecide TransactCommitPolicy = iota
	TransactCommitPolicy_TwoPhase
)

type TransactCommitPolicy int
