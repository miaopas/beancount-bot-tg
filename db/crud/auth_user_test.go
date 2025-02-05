package crud_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

func TestEnrichUserData(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	// nil message
	err = r.EnrichUserData(nil)
	if err == nil {
		t.Errorf("Expected error for nil message")
	}

	// User not already exists
	mock.
		ExpectQuery(`
			SELECT "tgUsername"
			FROM "auth::user"
			WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"tgUsername"}))

	mock.
		ExpectExec(`INSERT INTO "auth::user"`).
		WithArgs(chat.ID, "username").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = r.EnrichUserData(&tb.Message{Chat: chat, Sender: &tb.User{ID: chat.ID, Username: "username"}})
	if err != nil {
		t.Errorf("No error should be returned: %s", err.Error())
	}

	// User already exists, but username changed
	mock.
		ExpectQuery(`
			SELECT "tgUsername"
			FROM "auth::user"
			WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"tgUsername"}).AddRow("old_username"))

	mock.
		ExpectExec(`UPDATE "auth::user"`).
		WithArgs(chat.ID, "username").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = r.EnrichUserData(&tb.Message{Chat: chat, Sender: &tb.User{ID: chat.ID, Username: "username"}})
	if err != nil {
		t.Errorf("No error should be returned: %s", err.Error())
	}

	// User already exists, everything is fine
	err = r.EnrichUserData(&tb.Message{Chat: chat, Sender: &tb.User{ID: chat.ID, Username: "old_username"}})
	if err != nil {
		t.Errorf("No error should be returned: %s", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteUser(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.
		ExpectExec(`
			DELETE FROM "auth::user"
			WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = r.DeleteUser(&tb.Message{Chat: chat})
	if err != nil {
		t.Errorf("Expected no error for user deletion")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestEnrichUserDataErrors(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	crud.USER_CACHE[987] = &crud.UserCacheEntry{} // For test coverage: Old entry will be removed by userCachePrune

	// Error returned
	mock.
		ExpectQuery(`
			SELECT "tgUsername"
			FROM "auth::user"
			WHERE "tgChatId" = ?
		`).
		WithArgs(789).
		WillReturnError(fmt.Errorf("testing error behavior for EnrichUserData"))
	err = r.EnrichUserData(&tb.Message{Chat: &tb.Chat{ID: 789}, Sender: &tb.User{Username: "username2"}})
	if err == nil {
		t.Errorf("There should have been an error from the db query")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserGetCurrency(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}))
	cur := r.UserGetCurrency(&tb.Message{Chat: chat})
	if cur != helpers.DEFAULT_CURRENCY {
		t.Errorf("If no currency is given for user in db, use default currency")
	}

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TEST_CUR"))
	cur = r.UserGetCurrency(&tb.Message{Chat: chat})
	if cur != "TEST_CUR" {
		t.Errorf("If currency is given for user in db, that one should be used")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserIsAdmin(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(chat.ID, helpers.USERSET_ADM).
		WillReturnRows(sqlmock.NewRows([]string{"value"}))
	res := r.UserIsAdmin(&tb.Message{Chat: chat})
	if res {
		t.Errorf("User should not be admin")
	}

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(chat.ID, helpers.USERSET_ADM).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("false"))
	res = r.UserIsAdmin(&tb.Message{Chat: chat})
	if res {
		t.Errorf("User should not be admin")
	}

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(chat.ID, helpers.USERSET_ADM).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(true))
	res = r.UserIsAdmin(&tb.Message{Chat: chat})
	if !res {
		t.Errorf("User should be admin")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestIndividualsWithNotifications(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectQuery(`SELECT "tgChatId" FROM "auth::user"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"tgChatId"}))
	res := r.IndividualsWithNotifications("12345")
	if !helpers.ArraysEqual(res, []string{}) {
		t.Errorf("Some specific recipient should not be found: %v", res)
	}

	// Some specific recipient, but found
	mock.ExpectQuery(`SELECT "tgChatId" FROM "auth::user"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"tgChatId"}).AddRow(12345))
	res = r.IndividualsWithNotifications("12345")
	if !helpers.ArraysEqual(res, []string{"12345"}) {
		t.Errorf("Some specific recipient should be found: %v", res)
	}

	// All recipients from db
	mock.ExpectQuery(`SELECT "tgChatId" FROM "auth::user"`).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"tgChatId"}).AddRow(12345).AddRow(123456))
	res = r.IndividualsWithNotifications("")
	if !helpers.ArraysEqual(res, []string{"12345", "123456"}) {
		t.Errorf("All recipients should be returned from db: %v", res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserNotificationSetting(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectExec(`DELETE FROM "bot::notificationSchedule"`).
		WithArgs(12345).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "bot::notificationSchedule"`).
		WithArgs(12345, 3*24, 17).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = r.UserSetNotificationSetting(&tb.Message{Chat: &tb.Chat{ID: 12345}}, 3, 17)
	if err != nil {
		t.Errorf("Should not fail for setting notification setting")
	}

	mock.ExpectQuery(`SELECT "delayHours", "notificationHour" FROM "bot::notificationSchedule"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"delayHours", "notificationHour"}).AddRow(24*4, 18))
	daysDelay, hour, err := r.UserGetNotificationSetting(&tb.Message{Chat: &tb.Chat{ID: 12345}})
	if err != nil {
		t.Errorf("Should not fail for getting notification setting")
	}
	if daysDelay != 4 || hour != 18 {
		t.Errorf("Unexpected daysDelay (%d) or hour (%d)", daysDelay, hour)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserSetTag(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	message := &tb.Message{Chat: &tb.Chat{ID: 12345}}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_TAG).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "bot::userSetting"`).
		WithArgs(12345, helpers.USERSET_TAG, "TestTag").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err = r.UserSetTag(message, "TestTag")
	if err != nil {
		t.Errorf("Setting tag failed: %s", err.Error())
	}

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_TAG).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	r.UserSetTag(message, "")
	if err != nil {
		t.Errorf("Setting tag failed: %s", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserGetTag(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(9123, helpers.USERSET_TAG).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TEST_TAG"))
	tag := r.UserGetTag(&tb.Message{Chat: &tb.Chat{ID: 9123}})
	if tag != "TEST_TAG" {
		t.Errorf("TEST_TAG should have been returned as tag: %s", tag)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUserSetCurrency(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := crud.NewRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(9123, helpers.USERSET_CUR).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "bot::userSetting"`).
		WithArgs(9123, helpers.USERSET_CUR, "TEST_CUR_SPECIAL").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err = r.UserSetCurrency(&tb.Message{Chat: &tb.Chat{ID: 9123}}, "TEST_CUR_SPECIAL")
	if err != nil {
		t.Errorf("No error should have been thrown: %s", err.Error())
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
