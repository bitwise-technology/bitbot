package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

var (
	Token     = os.Getenv("BOT_TOKEN")
	ChannelID = os.Getenv("CHANNEL_ID")
	Message   = "@everyone - Não esqueçam de iniciar a contagem do tempo no clockify."
)

var ScheduleList = []string{
	"08:00",
	"13:00",
}

var done = make(chan bool)

func main() {
	// Create a new DiscordGo session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal("Error to create a DiscordGo session:", err)
		return
	}

	// Open connection with Discord
	err = dg.Open()
	if err != nil {
		log.Fatal("Error to open connection with Discord:", err)
		return
	}

	// Capture the interrupt signal to properly terminate the bot
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		<-sc
		dg.Close()
		close(done)
	}()

	// Start timer to send messages at fixed times
	go func() {
		for {
			now := time.Now()

			// Scroll through the list of scheduled times
			for _, schedule := range ScheduleList {
				scheduledTime, err := time.Parse("15:04", schedule)
				if err != nil {
					log.Fatal("Error parsing scheduled time:", err)
					continue
				}

				// Calculate next run based on scheduled time
				nextExecution := time.Date(now.Year(), now.Month(), now.Day(), scheduledTime.Hour(), scheduledTime.Minute(), 0, 0, now.Location())
				if nextExecution.Before(now) {
					nextExecution = nextExecution.Add(24 * time.Hour) // If it's past the scheduled time, add a day
				}

				// Wait until the scheduled time to send the message
				time.Sleep(nextExecution.Sub(now))

				// Find the channel object
				channel, err := dg.Channel(ChannelID)
				if err != nil {
					fmt.Println("Error finding channel:", err)
					continue
				}

				// Send the message to the specified channel
				_, err = dg.ChannelMessageSend(channel.ID, Message)
				if err != nil {
					fmt.Println("Error send message:", err)
					continue
				}
			}
		}
	}()

	<-done
}
