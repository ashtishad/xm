package common

const (
	ErrUnexpectedServer   = "unexpected server error occurred"
	ErrInvalidRequest     = "failed to validate request"
	ErrUnexpectedDatabase = "unexpected database error"
	ErrTXBegin            = "failed to begin transaction"
	ErrTXRollback         = "failed to rollback transaction"
	ErrTxCommit           = "failed to commit transaction"
	ErrIncorrectPassword  = "incorrect password"
)
