package bc

import "context"

// OpsIf is an interface that defines the operations that can be performed on the blockchain.
// Different implementations can choose to implement this interface.
type OpsIf interface {

	// MintDocTkn creates a new docTkn in some blockchain
	MintDocTkn(ctx context.Context, docId, docHash, ownerEmailHash string) (tknId string, err error)

	// VerifyDocTkn verifies the docHash, ownerEmail by retrieving the details using docTkn from some blockchain
	VerifyDocTkn(ctx context.Context, tknId, docHash, ownerEmailHash string) (err error)
}
