package cmd

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/yjwong/lark-cli/internal/api"
	"github.com/yjwong/lark-cli/internal/output"
)

var msgCmd = &cobra.Command{
	Use:   "msg",
	Short: "Message commands",
	Long:  "Retrieve and manage messages in Lark chats",
}

// --- msg history ---

var (
	msgHistoryChatID    string
	msgHistoryType      string
	msgHistoryStartTime string
	msgHistoryEndTime   string
	msgHistorySort      string
	msgHistoryLimit     int
)

var msgHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Get chat message history",
	Long: `Retrieve message history from a chat or thread.

Requires the bot to be in the group chat. For group chats, the app must have
the "Read all messages in associated group chat" permission scope.

Examples:
  lark msg history --chat-id oc_xxxxx
  lark msg history --chat-id oc_xxxxx --limit 50
  lark msg history --chat-id oc_xxxxx --start 1704067200 --end 1704153600
  lark msg history --chat-id oc_xxxxx --sort desc
  lark msg history --chat-id thread_xxxxx --type thread`,
	Run: func(cmd *cobra.Command, args []string) {
		if msgHistoryChatID == "" {
			output.Fatalf("VALIDATION_ERROR", "chat-id is required")
		}

		client := api.NewClient()

		// Build options
		opts := &api.ListMessagesOptions{}

		if msgHistoryStartTime != "" {
			opts.StartTime = parseTimeArg(msgHistoryStartTime)
		}
		if msgHistoryEndTime != "" {
			opts.EndTime = parseTimeArg(msgHistoryEndTime)
		}
		if msgHistorySort != "" {
			if msgHistorySort == "asc" {
				opts.SortType = "ByCreateTimeAsc"
			} else if msgHistorySort == "desc" {
				opts.SortType = "ByCreateTimeDesc"
			} else {
				output.Fatalf("VALIDATION_ERROR", "sort must be 'asc' or 'desc'")
			}
		}

		// Fetch messages with pagination
		var allMessages []api.Message
		var pageToken string
		hasMore := true
		remaining := msgHistoryLimit

		for hasMore {
			// Calculate page size
			pageSize := 50
			if remaining > 0 && remaining < pageSize {
				pageSize = remaining
			}
			opts.PageSize = pageSize
			opts.PageToken = pageToken

			messages, more, nextToken, err := client.ListMessages(msgHistoryType, msgHistoryChatID, opts)
			if err != nil {
				output.Fatal("API_ERROR", err)
			}

			allMessages = append(allMessages, messages...)
			hasMore = more
			pageToken = nextToken

			// Check limit
			if msgHistoryLimit > 0 {
				remaining = msgHistoryLimit - len(allMessages)
				if remaining <= 0 {
					break
				}
			}
		}

		// Trim to limit if needed
		if msgHistoryLimit > 0 && len(allMessages) > msgHistoryLimit {
			allMessages = allMessages[:msgHistoryLimit]
		}

		// Convert to output format
		outputMessages := make([]api.OutputMessage, len(allMessages))
		for i, m := range allMessages {
			outputMessages[i] = convertMessage(m)
		}

		result := api.OutputMessageList{
			Messages: outputMessages,
			Count:    len(outputMessages),
			ChatID:   msgHistoryChatID,
		}

		output.JSON(result)
	},
}

// parseTimeArg parses a time argument as either Unix timestamp or ISO 8601
func parseTimeArg(s string) string {
	// First try as Unix timestamp
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return s
	}

	// Try parsing as ISO 8601
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try without timezone
		t, err = time.Parse("2006-01-02T15:04:05", s)
		if err != nil {
			// Try date only
			t, err = time.Parse("2006-01-02", s)
			if err != nil {
				output.Fatalf("PARSE_ERROR", "invalid time format: %s (use Unix timestamp or ISO 8601)", s)
			}
		}
	}

	return strconv.FormatInt(t.Unix(), 10)
}

// convertMessage converts an API message to CLI output format
func convertMessage(m api.Message) api.OutputMessage {
	out := api.OutputMessage{
		MessageID:  m.MessageID,
		MsgType:    m.MsgType,
		CreateTime: formatMessageTime(m.CreateTime),
		IsReply:    m.RootID != "" || m.ParentID != "",
		ThreadID:   m.ThreadID,
		Deleted:    m.Deleted,
	}

	if m.Body != nil {
		out.Content = m.Body.Content
	}

	if m.Sender != nil {
		out.Sender = &api.OutputMessageSender{
			ID:   m.Sender.ID,
			Type: m.Sender.SenderType,
		}
	}

	if len(m.Mentions) > 0 {
		out.Mentions = make([]api.OutputMessageMention, len(m.Mentions))
		for i, mention := range m.Mentions {
			out.Mentions[i] = api.OutputMessageMention{
				Key:  mention.Key,
				ID:   mention.ID,
				Name: mention.Name,
			}
		}
	}

	return out
}

// formatMessageTime converts Unix milliseconds to ISO 8601
func formatMessageTime(ms string) string {
	if ms == "" {
		return ""
	}

	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return ms
	}

	t := time.UnixMilli(msInt)
	return t.Format(time.RFC3339)
}

// --- msg resource ---

var (
	msgResourceMessageID string
	msgResourceFileKey   string
	msgResourceType      string
	msgResourceOutput    string
)

var msgResourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Download a resource file from a message",
	Long: `Download resource files (images, videos, audios, files) from messages.

The file_key can be found in the message content JSON returned by 'lark msg history'.
For image messages, use --type image. For file, audio, and video messages, use --type file.

Examples:
  lark msg resource --message-id om_xxx --file-key img_v2_xxx --type image --output ./image.png
  lark msg resource --message-id om_xxx --file-key file_v2_xxx --type file --output ./video.mp4`,
	Run: func(cmd *cobra.Command, args []string) {
		if msgResourceMessageID == "" {
			output.Fatalf("VALIDATION_ERROR", "message-id is required")
		}
		if msgResourceFileKey == "" {
			output.Fatalf("VALIDATION_ERROR", "file-key is required")
		}
		if msgResourceType == "" {
			output.Fatalf("VALIDATION_ERROR", "type is required (image or file)")
		}
		if msgResourceType != "image" && msgResourceType != "file" {
			output.Fatalf("VALIDATION_ERROR", "type must be 'image' or 'file'")
		}
		if msgResourceOutput == "" {
			output.Fatalf("VALIDATION_ERROR", "output is required")
		}

		client := api.NewClient()

		// Download the resource
		body, contentType, err := client.GetMessageResource(msgResourceMessageID, msgResourceFileKey, msgResourceType)
		if err != nil {
			output.Fatal("API_ERROR", err)
		}
		defer body.Close()

		// Create output file
		outFile, err := os.Create(msgResourceOutput)
		if err != nil {
			output.Fatalf("FILE_ERROR", "failed to create output file: %v", err)
		}
		defer outFile.Close()

		// Copy data to file
		bytesWritten, err := io.Copy(outFile, body)
		if err != nil {
			output.Fatalf("FILE_ERROR", "failed to write file: %v", err)
		}

		// Output result
		result := map[string]interface{}{
			"success":       true,
			"message_id":    msgResourceMessageID,
			"file_key":      msgResourceFileKey,
			"output_path":   msgResourceOutput,
			"content_type":  contentType,
			"bytes_written": bytesWritten,
		}
		output.JSON(result)
	},
}

func init() {
	// msg history flags
	msgHistoryCmd.Flags().StringVar(&msgHistoryChatID, "chat-id", "", "Chat ID or thread ID (required)")
	msgHistoryCmd.Flags().StringVar(&msgHistoryType, "type", "chat", "Container type: 'chat' or 'thread'")
	msgHistoryCmd.Flags().StringVar(&msgHistoryStartTime, "start", "", "Start time (Unix timestamp or ISO 8601)")
	msgHistoryCmd.Flags().StringVar(&msgHistoryEndTime, "end", "", "End time (Unix timestamp or ISO 8601)")
	msgHistoryCmd.Flags().StringVar(&msgHistorySort, "sort", "", "Sort order: 'asc' or 'desc' (default: asc)")
	msgHistoryCmd.Flags().IntVar(&msgHistoryLimit, "limit", 0, "Maximum number of messages to retrieve (0 = no limit)")

	// msg resource flags
	msgResourceCmd.Flags().StringVar(&msgResourceMessageID, "message-id", "", "Message ID containing the resource (required)")
	msgResourceCmd.Flags().StringVar(&msgResourceFileKey, "file-key", "", "Resource file key from message content (required)")
	msgResourceCmd.Flags().StringVar(&msgResourceType, "type", "", "Resource type: 'image' or 'file' (required)")
	msgResourceCmd.Flags().StringVar(&msgResourceOutput, "output", "", "Output file path (required)")

	// Register subcommands
	msgCmd.AddCommand(msgHistoryCmd)
	msgCmd.AddCommand(msgResourceCmd)
}
