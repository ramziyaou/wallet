package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	"wallet/domain"
	"wallet/myerrors"
	"wallet/wallet/repository"

	"github.com/go-sql-driver/mysql"
)

type mySQLDBInterface struct {
	db *sql.DB
}

// GetLastAccountNo retrieves most recent account from DB
func (m *mySQLDBInterface) GetLastAccountNo() (string, error) {
	var lastAccountNo string
	if err := m.db.QueryRow("select `accountno` from wallets ORDER BY id DESC LIMIT 1").Scan(&lastAccountNo); err != nil {
		return "", err
	}
	if len(lastAccountNo) != 13 {
		return "", fmt.Errorf("Invalid length of account retrieved")
	}
	return lastAccountNo, nil
}

// InsertWallet inserts newly created wallet into DB
func (m *mySQLDBInterface) InsertWallet(account, IIN string) error {
	insForm, err := m.db.Prepare("insert into wallets (accountno, iin) values(?, ?)")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if _, err := insForm.Exec(account, IIN); err != nil {
		return err
	}
	return nil
}

// GetWallets retrieves wallets by given IIN
func (m *mySQLDBInterface) GetWallets(IIN string) ([]domain.Wallet, error) {
	rows, err := m.db.Query("SELECT accountno, id, ts, updated_at, amount FROM wallets WHERE iin = ?", IIN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var wallets []domain.Wallet
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var wallet domain.Wallet
		if err := rows.Scan(&wallet.AccountNo, &wallet.ID, &wallet.Ts, &wallet.UpdatedAt, &wallet.Amount); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

// TopUp implements account replenishment in DB
func (m *mySQLDBInterface) TopUp(account string, amt int) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	stmtWallet, err := tx.Prepare(`UPDATE wallets SET amount = amount + ? WHERE accountno = ?`)

	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmtWallet.Close()

	res, err := stmtWallet.Exec(amt, account)
	if err != nil {
		tx.Rollback()
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows == 0 {
		log.Println("ERROR|TopUp DB rows=0", err)
		return myerrors.ErrUpdateRows
	}

	stmtTx, err := tx.Prepare(`INSERT INTO transactions(transfer_type,from_acc,to_acc,amount) VALUES(?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmtTx.Close()

	res, err = stmtTx.Exec("topup", "", account, amt)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// GetAmount retrieves account amount
func (m *mySQLDBInterface) GetAmount(account string) (string, error) {
	var amount string
	if err := m.db.QueryRow("SELECT amount FROM wallets WHERE accountno = ?", account).Scan(&amount); err != nil {
		return "", err
	}
	return amount, nil
}

// SetZeroStatus sets transfer status back to status quo
func (m *mySQLDBInterface) SetZeroStatus(from, to string) error {
	res, err := m.db.Exec("UPDATE wallets SET transfer=0 WHERE accountno=? OR accountno=?", from, to)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 2 {
		log.Println("ERROR|Failed to update both rows, updated only", rows)
		return myerrors.ErrUpdateRows
	}
	return nil
}

// ConfirmIIN checks against IIN for requested account in DB
func (m *mySQLDBInterface) ConfirmIIN(IIN, account string) (bool, error) {
	var DBIIN string
	if err := m.db.QueryRow("SELECT iin FROM wallets WHERE accountno = ?", account).Scan(&DBIIN); err != nil {
		return false, err
	}
	return DBIIN == IIN, nil
}

// GetTransaction gets all transactions on given account
func (m *mySQLDBInterface) GetTransactions(account string) ([]domain.Transaction, error) {
	log.Println("ACCCCCC", account)
	rows, err := m.db.Query("SELECT * FROM transactions WHERE from_acc = ? OR to_acc = ?", account, account)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var transactions []domain.Transaction
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.Ts, &transaction.Type, &transaction.From, &transaction.To, &transaction.Amount); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	log.Println("transactions", transactions)
	return transactions, nil
}

// Transfer handles money transfer between accounts
func (m *mySQLDBInterface) Transfer(from, to, amt string) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	stmtFrom, err := tx.Prepare(`UPDATE wallets SET amount = amount - ? WHERE accountno = ?`)

	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmtFrom.Close()

	res, err := stmtFrom.Exec(amt, from)
	if err != nil {
		tx.Rollback()
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1690 {
			return myerrors.ErrInsufficientFunds
		}
		return err
	}

	stmtTo, err := tx.Prepare(`UPDATE wallets SET amount = amount + ? WHERE accountno = ?`)

	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmtTo.Close()

	res, err = stmtTo.Exec(amt, to)
	if err != nil {
		tx.Rollback()
		return err
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		tx.Rollback()
		return err
	}

	stmtTx, err := tx.Prepare(`INSERT INTO transactions(transfer_type,from_acc,to_acc,amount) VALUES(?,?,?,?)`)
	if err != nil {
		tx.Rollback()
		return err
	}

	defer stmtTx.Close()

	res, err = stmtTx.Exec("transfer", from, to, amt)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error inserting transaction")
		return err
	}
	fmt.Println("record added to transaction table successfully")
	fmt.Println(res)

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// UpdateTransferStatus sets status to 1 when transfer starts
func (m *mySQLDBInterface) UpdateTransferStatus(from, to string) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec("UPDATE wallets SET transfer=1 WHERE accountno=? OR accountno=? AND transfer=0", from, to)
	if err != nil {
		tx.Rollback()
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	fmt.Println(rows)
	if rows != 2 {
		tx.Rollback()
		log.Println("UpdateTransferStatus DB rows != 2 but", rows)
		return myerrors.ErrUpdateRows
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// GetWalletList gets all accounts under requested user and returns them in he form of string slice
func (m *mySQLDBInterface) GetWalletList(IIN string) ([]string, error) {
	rows, err := m.db.Query("SELECT accountno FROM wallets WHERE iin = ?", IIN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var wallets []string
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var wallet string
		if err := rows.Scan(&wallet); err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return wallets, nil
}

// NewMySQLDBInterace returns a new DB
func NewMySQLDBInterface(dbURL string) (repository.DBInterface, error) {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Println("ERROR|Error while opening DB:", err)
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 2)
	start := time.Now()
	for db.Ping() != nil {
		if time.Now().After(start.Add(time.Minute * 20)) {
			log.Println("ERROR|Failed to connect after 20 minutes")
			return nil, db.Ping()
		}
	}
	log.Println("DB Pong", db.Ping() == nil)

	return &mySQLDBInterface{db: db}, nil
}

// Close closes DB
func (m *mySQLDBInterface) Close() {
	m.db.Close()
}
