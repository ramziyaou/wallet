package mysql

import (
	"database/sql"
	"log"
	"testing"
	"wallet/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var w = &domain.Wallet{
	ID:        1,
	Ts:        "2021-12-31 19:36:36",
	UpdatedAt: "",
	AccountNo: "KZT0000000001",
	IIN:       "910815450350",
	Transfer:  0,
	Amount:    0,
}

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestLastAccountNo(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "select `accountno` from wallets ORDER BY id DESC LIMIT 1"

	rows := sqlmock.NewRows([]string{"accountno"}).
		AddRow(w.AccountNo)

	mock.ExpectQuery(query).WillReturnRows(rows)

	acc, err := repo.GetLastAccountNo()
	assert.NotNil(t, acc)
	assert.NoError(t, err)
}

func TestLastAccountErr(t *testing.T) {
	db, mock := NewMock()
	repo := &mySQLDBInterface{db}

	query := "select `accountno` from wallets ORDER BY id DESC LIMIT 1"

	rows := sqlmock.NewRows([]string{"accountno"}).
		AddRow("KZT000")

	mock.ExpectQuery(query).WillReturnRows(rows)

	acc, err := repo.GetLastAccountNo()
	db.Close()
	assert.NotNil(t, acc)
	assert.Error(t, err)
	//closed db
	acc, err = repo.GetLastAccountNo()
	if acc != "" {
		t.Errorf("Expected empty string, but got %s", acc)
	}
	assert.Error(t, err)

}

func TestInsertWallet(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}
	db.Begin()

	query := "insert into wallets (accountno, iin) values(?, ?)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs(w.AccountNo, w.IIN).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.InsertWallet(w.AccountNo, w.IIN)
	assert.NoError(t, err)
}

func TestInsertWalletErr(t *testing.T) {
	//closed DB
	db, mock := NewMock()
	db.Close()
	repo := &mySQLDBInterface{db}
	db.Begin()

	query := "insert into wallets (accountno, iin) values(?, ?)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs(w.AccountNo, w.IIN).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.InsertWallet(w.AccountNo, w.IIN)
	assert.Error(t, err)
}

func TestGetWallets(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT accountno, id, ts, updated_at, amount FROM wallets WHERE iin = ?"

	rows := sqlmock.NewRows([]string{"accountno", "id", "ts", "updated_at", "amount"}).
		AddRow(w.AccountNo, w.ID, w.Ts, w.UpdatedAt, w.Amount)

	mock.ExpectQuery(query).WithArgs(w.IIN).WillReturnRows(rows)
	wallets, err := repo.GetWallets(w.IIN)
	assert.NotNil(t, wallets)
	assert.NoError(t, err)
}

func TestGetWalletsErr(t *testing.T) {
	//closed DB
	db, mock := NewMock()
	db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT accountno, id, ts, updated_at, amount FROM wallets WHERE iin = ?"

	rows := sqlmock.NewRows([]string{"accountno", "id", "ts", "updated_at", "amount"}).
		AddRow(w.AccountNo, w.ID, w.Ts, w.UpdatedAt, w.Amount)

	mock.ExpectQuery(query).WithArgs(w.IIN).WillReturnRows(rows)
	wallets, err := repo.GetWallets(w.IIN)
	assert.Nil(t, wallets)
	assert.Error(t, err)
}

func TestGetAmount(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT amount FROM wallets WHERE accountno = ?"

	rows := sqlmock.NewRows([]string{"amount"}).
		AddRow(w.Amount)

	mock.ExpectQuery(query).WithArgs(w.AccountNo).WillReturnRows(rows)
	amt, err := repo.GetAmount(w.AccountNo)
	assert.NotNil(t, amt)
	assert.NoError(t, err)
}

func TestGetAmountErr(t *testing.T) {
	//closed DB
	db, mock := NewMock()
	db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT amount FROM wallets WHERE accountno = ?"

	rows := sqlmock.NewRows([]string{"amount"}).
		AddRow(w.Amount)

	mock.ExpectQuery(query).WithArgs(w.AccountNo).WillReturnRows(rows)
	amt, err := repo.GetAmount(w.AccountNo)
	if amt != "" {
		t.Errorf("Expected empty string, but got %s", amt)
	}
	assert.Error(t, err)
}

func TestConfirmIIN(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT iin FROM wallets WHERE accountno = ?"

	rows := sqlmock.NewRows([]string{"iin"}).
		AddRow(w.IIN)

	mock.ExpectQuery(query).WithArgs(w.AccountNo).WillReturnRows(rows)
	ok, err := repo.ConfirmIIN(w.IIN, w.AccountNo)
	if !ok {
		t.Error("Expected true, but got false")
	}
	assert.NoError(t, err)
}

func TestConfirmIINErr(t *testing.T) {
	db, mock := NewMock()
	repo := &mySQLDBInterface{db}

	query := "SELECT iin FROM wallets WHERE accountno = ?"

	rows := sqlmock.NewRows([]string{"iin"}).
		AddRow("980124450072")

	mock.ExpectQuery(query).WithArgs(w.AccountNo).WillReturnRows(rows)
	ok, err := repo.ConfirmIIN(w.IIN, w.AccountNo)
	db.Close()
	if ok {
		t.Error("Expected false, but got true")
	}
	assert.NoError(t, err)
	// closed DB
	ok, err = repo.ConfirmIIN(w.IIN, w.AccountNo)
	db.Close()
	if ok {
		t.Error("Expected false, but got true")
	}
	assert.Error(t, err)
}

func TestGetWalletList(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT accountno FROM wallets WHERE iin = ?"

	rows := sqlmock.NewRows([]string{"accountno"}).
		AddRow(w.AccountNo)

	mock.ExpectQuery(query).WithArgs(w.IIN).WillReturnRows(rows)
	wallets, err := repo.GetWalletList(w.IIN)
	assert.NotNil(t, wallets)
	assert.NoError(t, err)
}

func TestGetWalletListErr(t *testing.T) {
	// clsoed DB
	db, mock := NewMock()
	db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT accountno FROM wallets WHERE iin = ?"

	rows := sqlmock.NewRows([]string{"accountno"}).
		AddRow(w.AccountNo)

	mock.ExpectQuery(query).WithArgs(w.IIN).WillReturnRows(rows)
	wallets, err := repo.GetWalletList(w.IIN)
	assert.Nil(t, wallets)
	assert.Error(t, err)
}

var transaction = domain.Transaction{
	ID:     1,
	Type:   "topup",
	From:   "KZT0000000001",
	To:     "KZT0000000002",
	Amount: 123,
}

func TestGetTransactions(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "SELECT * FROM transactions WHERE from_acc = ? OR to_acc = ?"

	rows := sqlmock.NewRows([]string{"id", "ts", "type", "from", "to", "amount"}).
		AddRow(transaction.ID, transaction.Ts, transaction.Type, transaction.From, transaction.To, transaction.Amount)

	mock.ExpectQuery(query).WithArgs("KZT0000000001", "KZT0000000001").WillReturnRows(rows)
	txs, err := repo.GetTransactions("KZT0000000001")
	assert.NotNil(t, txs)
	assert.NoError(t, err)
}

func TestZeroStatusErr(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}
	query := "UPDATE wallets SET transfer=0 WHERE accountno=? OR accountno=?"

	mock.ExpectExec(query).WithArgs("KZT0000000001", "KZT0000000002").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetZeroStatus("KZT0000000001", "KZT0000000002")
	assert.Error(t, err)
}
