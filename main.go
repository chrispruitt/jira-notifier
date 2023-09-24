package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/slack-go/slack"
)

var (
	jiraClient *jira.Client
)

func main() {
	var (
		apiToken          string
		userEmail         string
		jql               string
		jiraUrl           string
		slackToken        string
		slackChannel      string
		slackMessageTitle string
	)

	// Read parameters from environment variables
	apiTokenEnv := os.Getenv("API_TOKEN")
	userEmailEnv := os.Getenv("USER_EMAIL")
	jqlEnv := os.Getenv("JQL")
	jiraUrlEnv := os.Getenv("JIRA_URL")
	slackTokenEnv := os.Getenv("SLACK_TOKEN")
	slackChannelEnv := os.Getenv("SLACK_CHANNEL")
	slackMessageTitleEnv := os.Getenv("SLACK_MSG_TITLE")

	// Override with command-line arguments if provided
	flag.StringVar(&apiToken, "api-token", apiTokenEnv, "Jira API token")
	flag.StringVar(&userEmail, "user", userEmailEnv, "Jira User Email")
	flag.StringVar(&jql, "jql", jqlEnv, "Jira Query Language (JQL) query")
	flag.StringVar(&jiraUrl, "jira-url", jiraUrlEnv, "Jira Instance URL")
	flag.StringVar(&slackToken, "slack-token", slackTokenEnv, "Slack App Token")
	flag.StringVar(&slackChannel, "slack-channel", slackChannelEnv, "Slack Channel")
	flag.StringVar(&slackMessageTitle, "slack-msg-title", slackMessageTitleEnv, "Slack Message Title")
	flag.Parse()

	if apiToken == "" || userEmail == "" || jiraUrl == "" || jql == "" || slackChannel == "" || slackToken == "" {
		fmt.Println("Please provide valid values for api-token, secret, jira-url, jql, slack-channel, slack-token")
		flag.Usage()
		os.Exit(1)
	}

	var err error
	jiraClient, err = createJiraClient(apiToken, jiraUrl, userEmail)
	if err != nil {
		log.Fatalf("Failed to create Jira client: %v", err)
	}

	issues, err := getIssues(jql)
	if err != nil {
		log.Fatalf("Failed to fetch Jira issues: %v", err)
	}

	if len(issues) > 0 {
		err = postToSlack(issues, slackToken, slackChannel, jiraUrl, jql, slackMessageTitle)
		if err != nil {
			log.Fatalf("Failed to post message: %v", err)
		}
	}
}

func createJiraClient(apiToken, jiraUrl, jiraEmail string) (*jira.Client, error) {
	tp := jira.BasicAuthTransport{
		Username: jiraEmail,
		Password: apiToken,
	}

	client, err := jira.NewClient(tp.Client(), jiraUrl)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getIssues(jql string) ([]jira.Issue, error) {
	issues, _, err := jiraClient.Issue.Search(jql, nil)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

func postToSlack(issues []jira.Issue, slackToken, channel, jiraUrl, jql, slackMessageTitle string) error {

	slackClient := slack.New(
		slackToken,
		// slack.OptionDebug(input.Debug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)

	msg := fmt.Sprintf("*%s*\n", slackMessageTitle)

	params := url.Values{}
	params.Add("jql", jql)
	msg += slackLink(fmt.Sprintf("%v Total Issues", len(issues)), fmt.Sprintf("%s/issues/?%s", jiraUrl, params.Encode()))

	_, msgTimestamp, err := slackClient.PostMessage(channel, slack.MsgOptionText(msg, false))
	if err != nil {
		return err
	}

	// Wait to ensure message has posted
	time.Sleep(500 * time.Millisecond)

	repyMsg := ""
	for _, issue := range issues {
		assignee := "Unassigned"
		if issue.Fields.Assignee != nil {
			assignee = issue.Fields.Assignee.DisplayName
		}

		repyMsg += fmt.Sprintf("%s - %s - %s", getAgeDays(time.Time(issue.Fields.Created)), assignee, slackLink(issue.Fields.Summary, fmt.Sprintf("%s/browse/%s", jiraUrl, issue.Key)))
		repyMsg += "\n"
	}

	_, _, _, err = slackClient.SendMessage(channel, slack.MsgOptionText(repyMsg, false), slack.MsgOptionTS(msgTimestamp))

	if err != nil {
		log.Fatalf("Failed to send thread reply: %v", err)
	}

	return nil
}

func slackLink(title string, url string) string {
	return fmt.Sprintf("<%s|%s>", url, title)
}

func getAgeDays(dt time.Time) string {
	now := time.Now()
	diff := now.Sub(dt)
	days := int(math.Round(diff.Hours() / 24))

	if days == 1 {
		return fmt.Sprintf("%d day", days)
	}

	return fmt.Sprintf("%d days", days)
}
