package dbtx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vposham/trustdoc/internal/db/sqlc/raw"
	"github.com/vposham/trustdoc/log"
	"go.uber.org/zap"
)

var errNoRec = fmt.Errorf("no records found")

type DocMeta struct {
	DocId          string
	OwnerEmail     string
	DocTitle       string
	DocDesc        string
	DocName        string
	OwnerFirstName string
	OwnerLastName  string
}

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
		DocHash:     "todo", // TODO
		DocMintedID: "todo", // TODO
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
func (store *Store) GetDocMeta(ctx context.Context, docId string) (DocMeta, error) {
	// TODO implement me
	panic("implement me")
}
