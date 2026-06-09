package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/bot"
	"github.com/boxingoctopus/snackmates/api/internal/seed"
	"github.com/spf13/cobra"
)

func main() {
	var (
		apiURL   string
		email    string
		password string
	)

	rootCmd := &cobra.Command{
		Use:   "bot",
		Short: "Artificial user for testing snack mate requests and messaging",
	}

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", bot.DefaultAPIURL, "SnackMates API base URL")
	rootCmd.PersistentFlags().StringVar(&email, "email", "bruno@snackmates.local", "bot account email")
	rootCmd.PersistentFlags().StringVar(&password, "password", seed.DefaultPassword, "bot account password")

	var (
		acceptFriends      bool
		replyChats         bool
		replyMessages      bool
		snagSnacks         bool
		snagDeliveryMethod string
		snagTrackingNumber string
		replyTemplate      string
		addFriends         []string
		sendChats          []string
		pollInterval       int
	)
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the bot and react to friend requests and messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, displayName, err := login(apiURL, email, password)
			if err != nil {
				return err
			}

			outgoing := make([]bot.OutgoingMessage, 0, len(sendChats))
			for _, raw := range sendChats {
				msg, err := bot.ParseOutgoingMessage(raw)
				if err != nil {
					return err
				}
				outgoing = append(outgoing, msg)
			}

			agent := bot.NewAgent(client, bot.Options{
				AcceptFriends:      acceptFriends,
				ReplyChats:         replyChats,
				ReplyMessages:      replyMessages,
				SnagSnacks:         snagSnacks,
				SnagDeliveryMethod: snagDeliveryMethod,
				SnagTrackingNumber: snagTrackingNumber,
				ReplyTemplate:      replyTemplate,
				AddFriends:         addFriends,
				SendChats:          outgoing,
				PollInterval:       pollIntervalSeconds(pollInterval),
				DisplayName:        displayName,
			})

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			log.Printf("bot running as %s (%s)", displayName, email)
			if err := agent.Run(ctx); err != nil && ctx.Err() == nil {
				return err
			}
			log.Printf("bot stopped")
			return nil
		},
	}
	runCmd.Flags().BoolVar(&acceptFriends, "accept-friends", true, "automatically accept incoming snack mate requests")
	runCmd.Flags().BoolVar(&replyChats, "reply-chats", true, "automatically reply to unread live chats")
	runCmd.Flags().BoolVar(&replyMessages, "reply-messages", true, "automatically reply to unread direct messages")
	runCmd.Flags().BoolVar(&snagSnacks, "snag-snacks", true, "automatically snag unsnagged items on snack mates' public wishlists")
	runCmd.Flags().StringVar(&snagDeliveryMethod, "snag-delivery", "in_person", "delivery method for snags: in_person or mail")
	runCmd.Flags().StringVar(&snagTrackingNumber, "snag-tracking", "", "tracking number when --snag-delivery=mail")
	runCmd.Flags().StringVar(&replyTemplate, "reply-template", "", "reply template; supports {display_name}, {from}, {body}")
	runCmd.Flags().StringArrayVar(&addFriends, "add-friend", nil, "send a snack mate request on startup (repeatable)")
	runCmd.Flags().StringArrayVar(&sendChats, "send-chat", nil, "send a live chat on startup as username:body (repeatable)")
	runCmd.Flags().IntVar(&pollInterval, "poll-interval", 5, "seconds between polls when the live notification stream is unavailable")

	requestCmd := &cobra.Command{
		Use:   "request-friend USERNAME",
		Short: "Send a snack mate request and exit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := login(apiURL, email, password)
			if err != nil {
				return err
			}
			return bot.NewAgent(client, bot.Options{}).SendFriendRequest(args[0])
		},
	}

	chatCmd := &cobra.Command{
		Use:   "chat USERNAME MESSAGE...",
		Short: "Send a live chat message to a snack mate and exit",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := login(apiURL, email, password)
			if err != nil {
				return err
			}
			return bot.NewAgent(client, bot.Options{}).SendChatToUser(args[0], strings.Join(args[1:], " "))
		},
	}

	snagCmd := &cobra.Command{
		Use:   "snag WISHLIST_SLUG ITEM_ID",
		Short: "Mark a wishlist item as snagged and exit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := login(apiURL, email, password)
			if err != nil {
				return err
			}
			delivery, _ := cmd.Flags().GetString("snag-delivery")
			tracking, _ := cmd.Flags().GetString("snag-tracking")
			return bot.NewAgent(client, bot.Options{
				SnagDeliveryMethod: delivery,
				SnagTrackingNumber: tracking,
			}).SnagWishlistItem(args[0], args[1])
		},
	}
	snagCmd.Flags().String("snag-delivery", "in_person", "delivery method: in_person or mail")
	snagCmd.Flags().String("snag-tracking", "", "tracking number when --snag-delivery=mail")

	rootCmd.AddCommand(runCmd, requestCmd, chatCmd, snagCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func login(apiURL, email, password string) (*bot.Client, string, error) {
	client := bot.NewClient(apiURL)
	if err := client.Login(email, password); err != nil {
		return nil, "", fmt.Errorf("login: %w", err)
	}

	me, err := client.Me()
	if err != nil {
		return nil, "", err
	}
	name := me.DisplayName
	if name == "" {
		name = me.Username
	}
	return client, name, nil
}

func pollIntervalSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
