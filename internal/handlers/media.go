package handlers

import (
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tormenta-bot/internal/assets"
)

// photoMsgTracker tracks which message IDs are photo messages
// so we know whether to use editMessageMedia or editMessageText
var (
	photoMsgTracker = map[int64]map[int]bool{} // chatID -> msgID -> hasPhoto
	photoMsgMu      sync.RWMutex
)

func markAsPhoto(chatID int64, msgID int) {
	photoMsgMu.Lock()
	defer photoMsgMu.Unlock()
	if photoMsgTracker[chatID] == nil {
		photoMsgTracker[chatID] = map[int]bool{}
	}
	photoMsgTracker[chatID][msgID] = true
}

func isPhotoMsg(chatID int64, msgID int) bool {
	photoMsgMu.RLock()
	defer photoMsgMu.RUnlock()
	if m, ok := photoMsgTracker[chatID]; ok {
		return m[msgID]
	}
	return false
}

// isNotModified returns true when Telegram rejects an edit because
// the new content is identical to the current one. This is harmless —
// the message already shows the correct state, no action needed.
func isNotModified(err error) bool {
	return err != nil && strings.Contains(err.Error(), "message is not modified")
}

// =============================================
// SEND PHOTO WITH CAPTION
// Returns the sent message ID
// =============================================

func sendPhoto(chatID int64, imageKey string, caption string, keyboard *tgbotapi.InlineKeyboardMarkup) int {
	if assets.Default == nil || !assets.Default.FileExists(imageKey) {
		// Fallback to text if image not available
		sendMsg(chatID, caption, keyboard)
		return 0
	}

	var photoConfig tgbotapi.PhotoConfig

	// Try cached file_id first (much faster)
	if fileID := assets.Default.GetFileID(imageKey); fileID != "" {
		photoConfig = tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
	} else {
		// Upload from disk
		path := assets.Default.GetPath(imageKey)
		photoConfig = tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(path))
	}

	photoConfig.Caption = caption
	photoConfig.ParseMode = "Markdown"
	if keyboard != nil {
		photoConfig.ReplyMarkup = keyboard
	}

	msg, err := Bot.Send(photoConfig)
	if err != nil {
		log.Printf("sendPhoto error for key=%s: %v — falling back to text", imageKey, err)
		sendMsg(chatID, caption, keyboard)
		return 0
	}

	// Cache the file_id from Telegram's response
	if len(msg.Photo) > 0 && assets.Default.GetFileID(imageKey) == "" {
		bestPhoto := msg.Photo[len(msg.Photo)-1] // largest size
		assets.Default.SetFileID(imageKey, bestPhoto.FileID)
		log.Printf("🖼️  Cached file_id for %s", imageKey)
	}

	markAsPhoto(chatID, msg.MessageID)
	return msg.MessageID
}

// =============================================
// EDIT PHOTO MESSAGE (change image + caption)
// =============================================

func editPhoto(chatID int64, msgID int, imageKey string, caption string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	if assets.Default == nil || !assets.Default.FileExists(imageKey) {
		editMsg(chatID, msgID, caption, keyboard)
		return
	}

	// If this message was originally a text message, we can't edit it as media
	if !isPhotoMsg(chatID, msgID) {
		Bot.Request(tgbotapi.NewDeleteMessage(chatID, msgID))
		sendPhoto(chatID, imageKey, caption, keyboard)
		return
	}

	var inputMedia tgbotapi.InputMediaPhoto

	if fileID := assets.Default.GetFileID(imageKey); fileID != "" {
		inputMedia = tgbotapi.NewInputMediaPhoto(tgbotapi.FileID(fileID))
	} else {
		path := assets.Default.GetPath(imageKey)
		inputMedia = tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(path))
	}
	inputMedia.Caption = caption
	inputMedia.ParseMode = "Markdown"

	editMediaConfig := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    chatID,
			MessageID: msgID,
		},
		Media: inputMedia,
	}
	if keyboard != nil {
		editMediaConfig.BaseEdit.ReplyMarkup = keyboard
	}

	msg, err := Bot.Send(editMediaConfig)
	if err != nil {
		if isNotModified(err) {
			return // conteúdo já está correto, nada a fazer
		}
		log.Printf("editPhoto error for key=%s: %v", imageKey, err)
		// Fallback apenas para erros reais (não "not modified")
		Bot.Request(tgbotapi.NewDeleteMessage(chatID, msgID))
		sendPhoto(chatID, imageKey, caption, keyboard)
		return
	}

	// Cache file_id if we got it back
	if len(msg.Photo) > 0 && assets.Default.GetFileID(imageKey) == "" {
		bestPhoto := msg.Photo[len(msg.Photo)-1]
		assets.Default.SetFileID(imageKey, bestPhoto.FileID)
	}
	markAsPhoto(chatID, msgID)
}

// =============================================
// EDIT ONLY CAPTION (keep same photo)
// =============================================

func editCaption(chatID int64, msgID int, caption string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	if !isPhotoMsg(chatID, msgID) {
		editMsg(chatID, msgID, caption, keyboard)
		return
	}

	editCaptionConfig := tgbotapi.EditMessageCaptionConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    chatID,
			MessageID: msgID,
		},
		Caption:   caption,
		ParseMode: "Markdown",
	}
	if keyboard != nil {
		editCaptionConfig.BaseEdit.ReplyMarkup = keyboard
	}

	if _, err := Bot.Send(editCaptionConfig); err != nil {
		if isNotModified(err) {
			return // conteúdo já está correto
		}
		log.Printf("editCaption error: %v", err)
		editMsg(chatID, msgID, caption, keyboard)
	}
}

// =============================================
// SMART EDIT: automatically picks best method
// imageKey can be "" to keep current image
// =============================================

func smartEdit(chatID int64, msgID int, imageKey string, caption string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	if imageKey == "" {
		editCaption(chatID, msgID, caption, keyboard)
	} else {
		editPhoto(chatID, msgID, imageKey, caption, keyboard)
	}
}

// =============================================
// SEND WITH PHOTO — sends message with image,
// tracks if callback was from a photo message
// =============================================

func sendOrEditPhoto(chatID int64, msgID int, isCallback bool, imageKey string, caption string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	if isCallback {
		editPhoto(chatID, msgID, imageKey, caption, keyboard)
	} else {
		sendPhoto(chatID, imageKey, caption, keyboard)
	}
}
