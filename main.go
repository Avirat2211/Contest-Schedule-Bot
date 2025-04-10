package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

type Contest struct {
	Id                  int    `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Frozen              bool   `json:"frozen"`
	Phase               string `json:"phase"`
	RelativeTimeSeconds int64  `json:"relativeTimeSeconds"`
	StartTimeSeconds    int64  `json:"startTimeSeconds"`
	DurationSeconds     int64  `json:"durationSeconds"`
}

type Response struct {
	Status string    `json:"status"`
	Result []Contest `json:"result"`
}

func formatDuration(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
func reverseContests(contests []Contest) []Contest {
	for i, j := 0, len(contests)-1; i < j; i, j = i+1, j-1 {
		contests[i], contests[j] = contests[j], contests[i]
	}
	return contests
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "!CF" {
		response, err := http.Get("https://codeforces.com/api/contest.list?gym=false")
		if err != nil {
			log.Fatalf("Error fetching API: %v", err)
			s.ChannelMessage(m.ChannelID, "Error fetching API: "+err.Error())
			return
		}
		defer response.Body.Close()
		var data Response
		if err = json.NewDecoder(response.Body).Decode(&data); err != nil {
			log.Fatalf("Error decoding JSON: %v", err)
			s.ChannelMessage(m.ChannelID, "Error in CF API: "+err.Error())
			return
		}
		if data.Status != "OK" {
			log.Fatalf("API returned non-OK status: %s", data.Status)
			s.ChannelMessage(m.ChannelID, "Error in CF API: "+data.Status)
			return
		}
		var upcoming []Contest
		fmt.Println("Upcoming Contests:")
		for _, contest := range data.Result {
			if contest.Phase == "BEFORE" {
				// fmt.Printf("ID: %d, Name: %s, Starts at: %d, Duration: %d seconds\n",
				// 	contest.Id, contest.Name, contest.StartTimeSeconds, contest.DurationSeconds)
				upcoming = append(upcoming, contest)
			}
		}
		reverseContests(upcoming)
		ist, _ := time.LoadLocation("Asia/Kolkata")
		message := "**Upcoming Codeforces Contests:**\n\n"
		for _, contest := range upcoming {
			// start := time.Unix(contest.StartTimeSeconds, 0).UTC()
			start := time.Unix(contest.StartTimeSeconds, 0).In(ist)
			message += fmt.Sprintf(
				"• **%s**\n  🕒 %s IST\n  ⏳ Duration: %s\n  🔗 https://codeforces.com/contests/%d\n\n",
				contest.Name,
				start.Format("02 Jan 2006 15:04"),
				formatDuration(contest.DurationSeconds),
				contest.Id,
			)
		}

		s.ChannelMessageSend(m.ChannelID, message)
	}
	if m.Content == "!LC" {

	}
}
func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
