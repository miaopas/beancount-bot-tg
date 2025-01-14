package crud_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

func TestHealthGetLogs(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	// TODO: Don't get query in here to parse (regex issues)
	mock.ExpectQuery("").
		WithArgs(helpers.ERROR, helpers.WARN, 24).
		WillReturnRows(sqlmock.NewRows([]string{"level", "c"}).
			AddRow(helpers.ERROR, 17))

	errors, warnings, err := r.HealthGetLogs(24)
	if err != nil {
		t.Errorf("Should not fail for getting health logs count: %s", err.Error())
	}
	if errors != 17 || warnings != 0 {
		t.Errorf("Unexpected logs count: errors: %d != %d or warnings: %d != %d", errors, 17, warnings, 0)
	}

	// TODO: Don't get query in here to parse (regex issues)
	mock.ExpectQuery("").
		WithArgs(helpers.ERROR, helpers.WARN, 2).
		WillReturnRows(sqlmock.NewRows([]string{"level", "c"}).
			AddRow(helpers.ERROR, 17).
			AddRow(helpers.WARN, 35))

	errors, warnings, err = r.HealthGetLogs(2)
	if err != nil {
		t.Errorf("Should not fail for getting health logs count: %s", err.Error())
	}
	if errors != 17 || warnings != 35 {
		t.Errorf("Unexpected logs count: errors: %d != %d or warnings: %d != %d", errors, 17, warnings, 35)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestHealthGetTransactions(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	// TODO: Don't get query in here to parse (regex issues)
	mock.ExpectQuery("").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"archived", "c"}).
			AddRow(true, 22).
			AddRow(false, 33))

	open, archived, err := r.HealthGetTransactions()
	if err != nil {
		t.Errorf("Should not fail for getting health transactions count: %s", err.Error())
	}
	if open != 33 || archived != 22 {
		t.Errorf("Unexpected transactions count: open: %d != %d or archived: %d != %d", open, 33, archived, 22)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestHealthGetUserCount(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	// TODO: Don't get query in here to parse (regex issues)
	mock.ExpectQuery("").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"c"}).
			AddRow(3184))

	count, err := r.HealthGetUserCount()
	if err != nil {
		t.Errorf("Should not fail for getting health transactions count: %s", err.Error())
	}
	if count != 3184 {
		t.Errorf("Unexpected users count: %d != %d", count, 3184)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestHealthGetCacheStats(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	// TODO: Don't get query in here to parse (regex issues)
	mock.ExpectQuery("").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"type", "c"}).
			AddRow("account:to", 7).
			AddRow("account:from", 9).
			AddRow("description:", 15).
			AddRow("blahbla1:", 21).
			AddRow("blahbla2:", 3))

	to, from, desc, other, err := r.HealthGetCacheStats()
	if err != nil {
		t.Errorf("Should not fail for getting health transactions count: %s", err.Error())
	}
	if to != 7 || from != 9 || desc != 15 || other != 24 {
		t.Errorf("Unexpected cache stats: %d != %d || %d != %d || %d != %d || %d != %d", to, 7, from, 9, desc, 15, other, 24)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestHealthGetUsersActiveCounts(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectQuery("").
		WithArgs(72).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).
			AddRow(4))

	count, err := r.HealthGetUsersActiveCounts(3 * 24)
	if err != nil {
		t.Errorf("Should not fail for getting active user count: %s", err.Error())
	}
	if count != 4 {
		t.Errorf("Unexpected active user count: %d != %d", count, 4)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
