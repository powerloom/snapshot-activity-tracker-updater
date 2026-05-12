package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Slack epoch summary webhook (same incoming-webhook pattern as snapshotter-lite-local-collector).

type slackWebhookMessage struct {
	Text        string           `json:"text,omitempty"`
	Username    string           `json:"username,omitempty"`
	IconEmoji   string           `json:"icon_emoji,omitempty"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

type slackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Fields    []slackField `json:"fields,omitempty"`
	Footer    string       `json:"footer,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func slackEpochSummaryEvery() int {
	v := strings.TrimSpace(os.Getenv("SLACK_EPOCH_SUMMARY_EVERY_EPOCHS"))
	if v == "" {
		return 30
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 30
	}
	return n
}

func slackWebhookURL() string {
	if u := strings.TrimSpace(os.Getenv("SLACK_WEBHOOK_URL")); u != "" {
		return u
	}
	return strings.TrimSpace(os.Getenv("ACTIVITY_TRACKER_SLACK_WEBHOOK_URL"))
}

func maybeSendSlackEpochSummary(ctx context.Context, tally *TallyDump, currentDay string) {
	url := slackWebhookURL()
	if url == "" || tally == nil {
		return
	}
	every := slackEpochSummaryEvery()
	if every == 0 {
		return
	}
	if tally.EpochID%uint64(every) != 0 {
		return
	}

	activeSlots := 0
	for _, c := range tally.SubmissionCounts {
		if c > 0 {
			activeSlots++
		}
	}

	withCID := 0
	withoutCID := 0
	for _, vs := range tally.ValidatorSummaries {
		if vs.HasBatchCID {
			withCID++
		} else {
			withoutCID++
		}
	}
	if len(tally.ValidatorSummaries) == 0 {
		withCID = len(tally.ValidatorBatchCIDs)
	}

	dayStr := currentDay
	if dayStr == "" {
		dayStr = "—"
	}

	fields := []slackField{
		{Title: "Epoch", Value: fmt.Sprintf("%d", tally.EpochID), Short: true},
		{Title: "Protocol day", Value: dayStr, Short: true},
		{Title: "Active slots (count>0)", Value: fmt.Sprintf("%d", activeSlots), Short: true},
		{Title: "Slots tracked", Value: fmt.Sprintf("%d", len(tally.SubmissionCounts)), Short: true},
		{Title: "Eligible nodes (epoch)", Value: fmt.Sprintf("%d", tally.EligibleNodesCount), Short: true},
		{Title: "Validators (batches)", Value: fmt.Sprintf("%d", tally.TotalValidators), Short: true},
		{Title: "Aggregated projects", Value: fmt.Sprintf("%d", len(tally.AggregatedProjects)), Short: true},
		{Title: "Validators w/ batch CID", Value: fmt.Sprintf("%d", withCID), Short: true},
		{Title: "Validators w/o batch CID", Value: fmt.Sprintf("%d", withoutCID), Short: true},
		{Title: "Data market", Value: shortenAddr(tally.DataMarket), Short: false},
	}

	title := fmt.Sprintf("DSV activity tracker · epoch %d", tally.EpochID)
	username := os.Getenv("SLACK_USERNAME")
	if username == "" {
		username = "Activity Tracker"
	}

	msg := slackWebhookMessage{
		Username:  username,
		IconEmoji: ":bar_chart:",
		Attachments: []slackAttachment{
			{
				Color:     "good",
				Title:     title,
				Fields:    fields,
				Footer:    "snapshot-activity-tracker-updater",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("slack epoch summary: marshal: %v", err)
		return
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		log.Printf("slack epoch summary: request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("slack epoch summary: post: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("slack epoch summary: webhook status %d", resp.StatusCode)
	}
}

func shortenAddr(a string) string {
	a = strings.TrimSpace(a)
	if len(a) <= 14 {
		return a
	}
	return a[:6] + "…" + a[len(a)-4:]
}
