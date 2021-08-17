package protocol

const transactionStatusCodeValid = uint32(0)
const transactionStatusCodeInvalidEmpty = uint32(1)

// Transaction is a generic transaction type
type Transaction struct {
	raw []byte
}

// TransactionFromRaw initializes a new Transaction given its wire representation
func TransactionFromRaw(tx []byte) (*Transaction, error) {
	return &Transaction{
		raw: tx,
	}, nil
}

func (tx *Transaction) calculateGas() int64 {
	// TODO
	return int64(0)
}

func (tx *Transaction) isValid() (code uint32) {
	if tx == nil || len(tx.raw) == 0 {
		return transactionStatusCodeInvalidEmpty
	}

	// TODO: check transaction format
	return transactionStatusCodeValid
}
