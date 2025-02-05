# Scenario tests for beancount-bot-tg

## Setup GitHub Action

A few prerequisites are required for the scenario tests to run effectively.

- You need to create a chat with the bot you would like to test.
- You need to get a Telegram app id and hash from https://me.telegram.org
- You need to store the values in [GitHub Actions secrets](https://github.com/LucaBernstein/beancount-bot-tg/settings/secrets/actions)

```bash
# Installing dependencies
pip install -r requirements.txt
# Creating authentication session: Fill in all details queried from you.
python authenticate.py
# Store the output of the following command as Actions secret 'SCENARIO_TG_ANON_SESSION'
cat anon.session | base64
# The reverse is easiest if encoded string is pasted into file and decoded from there:
#  cat anon.session.base64 | base64 -d > anon.session

# Next, create and fill the secrets 'SCENARIO_TG_API_ID', 'SCENARIO_TG_API_HASH' and 'SCENARIO_TG_CHAT_ID'
```

## Running the tests locally

```bash
# For acquiring authentication, see above
env $(cat .env | xargs) behave
```