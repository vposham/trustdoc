package dbtx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/vposham/trustdoc/internal/db/sqlc/raw"
	"github.com/vposham/trustdoc/log"
)

// DocMeta holds all the metadata for a document in postgres
type DocMeta struct {
	DocId          string `json:"docId,omitempty"`
	OwnerEmail     string `json:"ownerEmail,omitempty"`
	DocTitle       string `json:"docTitle,omitempty"`
	DocDesc        string `json:"docDesc,omitempty"`
	DocName        string `json:"docName,omitempty"`
	DocHash        string `json:"docHash,omitempty"`
	BcTknId        string `json:"bcTknId,omitempty"`
	OwnerFirstName string `json:"ownerFirstName,omitempty"`
	OwnerLastName  string `json:"ownerLastName,omitempty"`
}

// SaveDocMeta saves the document meta in postgres db with a retry based pessimistic locking transaction
func (store *Store) SaveDocMeta(ctx context.Context, in DocMeta) error {
	logger := log.GetLogger(ctx)
	logger.Info("started db tx for saving document meta", zap.Any("docId", in.DocId))
	return store.execTxWithRetry(ctx, func(queries Queries) error {
		u, exists, err := chkUsrExists(ctx, queries, in.OwnerEmail)
		if err != nil {
			return err
		}
		if !exists {
			u, err = createUser(ctx, queries, in.OwnerEmail, in.OwnerFirstName, in.OwnerLastName)
			if err != nil {
				return err
			}
		}
		return saveDocMeta(ctx, queries, in, u)
	})
}

// GetDocMetaByHash gets the document meta from postgres db with a retry based pessimistic locking transaction
func (store *Store) GetDocMetaByHash(ctx context.Context, docHash string) (DocMeta, error) {
	logger := log.GetLogger(ctx)
	logger.Info("started db tx for get document meta by hash", zap.Any("docHash", docHash))
	var m DocMeta
	err := store.execTxWithRetry(ctx, func(queries Queries) error {
		doc, err := queries.GetDocByHash(ctx, docHash)
		if err != nil {
			return err
		}
		u, err := queries.GetUserById(ctx, doc.UserID)
		if err != nil {
			return err
		}
		m = DocMeta{
			DocId:          doc.DocID,
			OwnerEmail:     u.EmailID,
			DocTitle:       doc.Title,
			DocDesc:        doc.Description.String,
			DocName:        doc.FileName,
			DocHash:        doc.DocHash,
			BcTknId:        doc.DocMintedID,
			OwnerFirstName: u.FirstName,
			OwnerLastName:  u.LastName,
		}
		return nil
	})
	return m, err
}

func chkUsrExists(ctx context.Context, queries Queries, email string) (u *raw.User, exists bool, err error) {
	logger := log.GetLogger(ctx)
	logger.Info("checking if user exists", zap.String("email", email))
	exists = false
	user, err := queries.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Error("user doesnt exist", zap.String("email", email))
			err = nil // set to nil to indicate no error
			return
		}
		logger.Error("failed to chkUsrExists", zap.String("email", email), zap.Error(err))
		err = fmt.Errorf("failed to chkUsrExists - %w", err)
		return
	}
	logger.Info("user exists", zap.String("email", email))
	exists = true
	u = &user
	return
}

func createUser(ctx context.Context, queries Queries, emailId, firstName, lastName string) (u *raw.User, err error) {
	logger := log.GetLogger(ctx)
	logger.Info("creating a new user", zap.String("email", emailId))
	arg := raw.AddUserParams{
		EmailID:   emailId,
		FirstName: firstName,
		LastName:  lastName,
		Status:    raw.UserTypeACTIVE,
	}
	user, err := queries.AddUser(ctx, arg)
	if err != nil {
		logger.Error("failed to createUser", zap.String("email", emailId), zap.Error(err))
		err = fmt.Errorf("failed to createUser - %w", err)
		return
	}
	u = &user
	return
}

func saveDocMeta(ctx context.Context, queries Queries, in DocMeta, u *raw.User) error {
	logger := log.GetLogger(ctx)
	logger.Info("saving document meta", zap.String("docId", in.DocId))
	arg := raw.AddDocParams{
		DocID:       in.DocId,
		Title:       in.DocTitle,
		Description: sql.NullString{String: in.DocDesc, Valid: true},
		FileName:    in.DocName,
		DocHash:     in.DocHash,
		DocMintedID: in.BcTknId,
		UserID:      u.ID,
	}
	_, err := queries.AddDoc(ctx, arg)
	if err != nil {
		logger.Error("failed to saveDocMeta", zap.String("docId", in.DocId), zap.Error(err))
		err = fmt.Errorf("failed to saveDocMeta - %w", err)
		return err
	}
	return nil
}
