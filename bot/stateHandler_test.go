package bot_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	tb "gopkg.in/telebot.v3"
)

func TestStateClearing(t *testing.T) {
	message := &tb.Message{Chat: &tb.Chat{ID: 24}}
	stateHandler := bot.NewStateHandler()

	stateHandler.SimpleTx(message, "")
	state := stateHandler.GetTx(message)
	if state == nil {
		t.Errorf("State from StateHandler before clearing was wrong, got: nil, want: not nil.")
	}

	stateHandler.Clear(message)
	if stateHandler.GetTx(message) != nil {
		t.Errorf("State from StateHandler after clearing was wrong, got: not nil, want: nil.")
	}
}
