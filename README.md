# jira-notifier

Jira Notifier is a simple cli that takes a jira jql query and posts the found issues to slack.

## Usage

```text
› jira-notifier --help

  -api-token string
        Jira API token
  -jira-url string
        Jira Instance URL
  -jql string
        Jira Query Language (JQL) query
  -slack-channel string
        Slack Channel
  -slack-msg-title string
        Slack Message Title
  -slack-token string
        Slack App Token
  -user string
        Jira User Email
```

### Docker

```sh
docker run -it \
     -e API_TOKEN="XXXX" \
    -e USER_EMAIL="XXXX" \
    -e JQL="resolution = Unresolved AND issuetype = Incident ORDER BY \"Time to resolution\" ASC" \
    -e JIRA_URL="https://myorg.atlassian.net" \
    -e SLACK_CHANNEL="CHANNEL_ID_XXXXX" \
    -e SLACK_TOKEN="XXXX" \
    -e SLACK_MSG_TITLE="JIRA Issue Digest" \
    chrispruitt/jira-notifier:latest
```
