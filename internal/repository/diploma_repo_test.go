package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/diasoft/diplomaverify/internal/crypto"
	"github.com/diasoft/diplomaverify/internal/models"
	"github.com/diasoft/diplomaverify/internal/repository"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
)

func TestDiplomaRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	cipher, _ := crypto.NewAESCipher("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	repo := repository.NewDiplomaRepository(mock, cipher)

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO diplomas").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), "CS", 2024, pgxmock.AnyArg(), "", pgxmock.AnyArg(), pgxmock.AnyArg(), "active").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow("123", time.Now()))
	mock.ExpectCommit()

	dto := models.CreateDiplomaDTO{
		UniversityID: "uni-1", SerialNumber: "DIP123", Specialty: "CS", IssueYear: 2024, FIOFull: "Test User",
	}
	_, err = repo.Create(context.Background(), dto, "hash", "sig")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}