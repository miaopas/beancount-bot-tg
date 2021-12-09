package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bc *BotController) configHandler(m *tb.Message) {
	sc := h.MakeSubcommandHandler("/"+CMD_CONFIG, true)
	sc.
		Add("currency", bc.configHandleCurrency).
		Add("tag", bc.configHandleTag).
		Add("notify", bc.configHandleNotification).
		Add("about", bc.configHandleAbout)
	err := sc.Handle(m)
	if err != nil {
		bc.configHelp(m, nil)
	}
}

func (bc *BotController) configHelp(m *tb.Message, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg += fmt.Sprintf("Error executing your command: %s\n\n", err.Error())
	}
	tz, _ := time.Now().Zone()
	_, err = bc.Bot.Send(m.Sender, errorMsg+fmt.Sprintf("Usage help for /%s:\n\n"+
		"/%s currency <c> - Change default currency"+
		"\n/%s about - Display the version this bot is running on"+
		"\n\nTags will be added to each new transaction with a '#':\n"+
		"\n/%s tag - Get currently set tag"+
		"\n/%s tag off - Turn off tag"+
		"\n/%s tag <name> - Set tag to apply to new transactions, e.g. when on vacation"+
		"\n\nCreate a schedule to be notified of open transactions (i.e. not archived or deleted):\n"+
		"\n/%s notify - Get current notification status"+
		"\n/%s notify off - Disable reminder notifications"+
		"\n/%s notify <delay> <hour> - Notify of open transaction after <delay> days at <hour> of the day (%s)",
		CMD_CONFIG, CMD_CONFIG, CMD_CONFIG, CMD_CONFIG, CMD_CONFIG, CMD_CONFIG, CMD_CONFIG, CMD_CONFIG, CMD_CONFIG, tz))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) configHandleCurrency(m *tb.Message, params ...string) {
	currency := bc.Repo.UserGetCurrency(m)
	if len(params) == 0 { // 0 params: GET currency
		// Return currently set currency
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Your current currency is set to '%s'. To change it add the new currency to use to the command like this: '/%s currency EUR'.", currency, CMD_CONFIG))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if len(params) > 1 { // 2 or more params: too many
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	// Set new currency
	newCurrency := params[0]
	err := bc.Repo.UserSetCurrency(m, newCurrency)
	if err != nil {
		_, err = bc.Bot.Send(m.Sender, "An error ocurred saving your currency preference: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("Changed default currency for all future transactions from '%s' to '%s'.", currency, newCurrency))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) configHandleTag(m *tb.Message, params ...string) {
	if len(params) == 0 {
		// GET tag
		tag := bc.Repo.UserGetTag(m)
		if tag != "" {
			_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("All new transactions automatically get the tag #%s added (vacation mode enabled)", tag))
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		} else {
			_, err := bc.Bot.Send(m.Sender, "No tags are currently added to new transactions (vacation mode disabled).")
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		}
		return
	} else if len(params) > 1 { // Only 0 or 1 allowed
		bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
		return
	}
	if params[0] == "off" {
		// DELETE tag
		bc.Repo.UserSetTag(m, "")
		_, err := bc.Bot.Send(m.Sender, "Disabled automatically set tags on new transactions")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	// SET tag
	tag := strings.TrimPrefix(params[0], "#")
	err := bc.Repo.UserSetTag(m, tag)
	if err != nil {
		_, err = bc.Bot.Send(m.Sender, "An error ocurred saving the tag: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("From now on all new transactions automatically get the tag #%s added (vacation mode enabled)", tag))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) configHandleNotification(m *tb.Message, params ...string) {
	var tz, _ = time.Now().Zone()
	if len(params) == 0 {
		// GET schedule
		daysDelay, hour, err := bc.Repo.UserGetNotificationSetting(m)
		if err != nil {
			bc.configHelp(m, fmt.Errorf("an application error occurred while retrieving user information from database"))
			return
		}
		if daysDelay < 0 {
			_, err = bc.Bot.Send(m.Sender, "Notifications are disabled for open transactions.")
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
			return
		}
		plural_s := "s"
		if daysDelay == 1 {
			plural_s = ""
		}
		_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("The bot will notify you daily at hour %d (%s) if transactions are open for more than %d day%s", hour, tz, daysDelay, plural_s))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	} else if len(params) == 1 {
		// DELETE schedule
		if params[0] == "off" {
			err := bc.Repo.UserSetNotificationSetting(m, -1, -1)
			if err != nil {
				bc.configHelp(m, fmt.Errorf("error setting notification schedule: %s", err.Error()))
				return
			}
			_, err = bc.Bot.Send(m.Sender, "Successfully disabled notifications for open transactions.")
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
			return
		}
		bc.configHelp(m, fmt.Errorf("invalid parameters"))
		return
	} else if len(params) == 2 {
		// SET schedule
		daysDelay, err := strconv.Atoi(params[0])
		if err != nil {
			bc.configHelp(m, fmt.Errorf("error converting daysDelay to number: %s: %s", params[0], err.Error()))
			return
		}
		hour, err := strconv.Atoi(params[1])
		if err != nil {
			bc.configHelp(m, fmt.Errorf("error converting hour to number: %s: %s", params[1], err.Error()))
			return
		}
		err = bc.Repo.UserSetNotificationSetting(m, daysDelay, hour)
		if err != nil {
			bc.configHelp(m, fmt.Errorf("error setting notification schedule: %s", err.Error()))
		}
		bc.configHandleNotification(m) // Recursively call with zero params --> GET
		return
	}
	bc.configHelp(m, fmt.Errorf("invalid amount of parameters specified"))
}

func (bc *BotController) configHandleAbout(m *tb.Message, params ...string) {
	if len(params) > 0 {
		bc.configHelp(m, fmt.Errorf("no parameters expected"))
		return
	}
	version := os.Getenv("VERSION")
	versionLink := "https://github.com/LucaBernstein/beancount-bot-tg/releases/"
	if strings.HasPrefix(version, "v") {
		versionLink += "tag/" + version
	}
	if version == "" {
		version = "not specified"
	}
	_, err := bc.Bot.Send(m.Sender, fmt.Sprintf(`Version information about [LucaBernstein/beancount\-bot\-tg](https://github.com/LucaBernstein/beancount\-bot\-tg)

Version: [%s](%s)`,
		version,
		strings.ReplaceAll(versionLink, "-", "\\-"),
	), tb.ModeMarkdownV2)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}
