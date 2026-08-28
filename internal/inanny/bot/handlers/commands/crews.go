package command

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	crews "github.com/holeyko/innany-tgbot/internal/inanny/features/crews"
)

type CrewCommandHandler struct {
	CommandHandler
}

func (handler CrewCommandHandler) Handle(bot *tgbot.BotAPI, update *tgbot.Update) error {
	args := strings.Fields(update.Message.CommandArguments())
	if len(args) == 0 {
		return errors.New("Usage: /crew <name>")
	}

	switch args[0] {
	case "create":
		return handleCreateCrew(bot, update, args)
	case "delete":
		return handleDeleteCrew(bot, update, args)
	default:
		if len(args) == 1 {
			return handleMentionCrew(bot, update, args[0])
		}
		return handleChangeCrewMember(bot, update, args)
	}
}

func handleCreateCrew(bot *tgbot.BotAPI, update *tgbot.Update, args []string) error {
	if len(args) < 3 {
		return errors.New("Usage: /crew create <name> <username>")
	}

	name := crews.NormalizeName(args[1])
	if name == "create" || name == "delete" {
		return fmt.Errorf("Crew name %q is reserved", name)
	}
	if err := crews.ValidateName(name); err != nil {
		return err
	}

	if update.Message.From == nil {
		return errors.New("A user account is required to create a crew")
	}
	creatorUsername := crews.NormalizeUsername(update.Message.From.UserName)
	if err := crews.ValidateUsername(creatorUsername); err != nil {
		return errors.New("You must have a Telegram username to create a crew")
	}

	members, err := parseCrewMembers(strings.Join(args[2:], " "), creatorUsername)
	if err != nil {
		return err
	}

	if _, err := crews.GetCrewByChatAndName(update.Message.Chat.ID, name); err == nil {
		return fmt.Errorf("Crew %q already exists in this chat", name)
	} else if !errors.Is(err, crews.ErrCrewNotFound) {
		return err
	}

	_, err = crews.CreateCrew(crews.CreateCrewDto{
		ChatID: update.Message.Chat.ID,
		Creator: crews.UserDto{
			TelegramLogin: creatorUsername,
			FirstName:     update.Message.From.FirstName,
			LastName:      update.Message.From.LastName,
		},
		Name:    name,
		Members: members,
	})
	if err != nil {
		return err
	}

	return sendCrewReply(bot, update, fmt.Sprintf("Crew %q was created", name))
}

func handleDeleteCrew(bot *tgbot.BotAPI, update *tgbot.Update, args []string) error {
	if len(args) != 2 {
		return errors.New("Usage: /crew delete <name>")
	}

	name := crews.NormalizeName(args[1])
	crew, err := getCrew(update, name)
	if err != nil {
		return err
	}
	if err := requireCrewManager(bot, update, crew); err != nil {
		return err
	}

	deleted, err := crews.DeleteCrew(crew.ID, update.Message.Chat.ID)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("Crew %q doesn't exist in this chat", name)
	}

	return sendCrewReply(bot, update, fmt.Sprintf("Crew %q was deleted", name))
}

func handleChangeCrewMember(bot *tgbot.BotAPI, update *tgbot.Update, args []string) error {
	if len(args) != 3 || (args[1] != "add" && args[1] != "delete") {
		return errors.New("Usage: /crew <name> add|delete <username>")
	}

	name := crews.NormalizeName(args[0])
	crew, err := getCrew(update, name)
	if err != nil {
		return err
	}
	if err := requireCrewManager(bot, update, crew); err != nil {
		return err
	}

	username := crews.NormalizeUsername(args[2])
	if err := crews.ValidateUsername(username); err != nil {
		return err
	}

	var changed bool
	if args[1] == "add" {
		changed, err = crews.AddCrewMember(crew.ID, username)
		if err == nil && !changed {
			return fmt.Errorf("@%s is already a member of crew %q", username, name)
		}
	} else {
		changed, err = crews.DeleteCrewMember(crew.ID, username)
		if err == nil && !changed {
			return fmt.Errorf("@%s is not a member of crew %q", username, name)
		}
	}
	if err != nil {
		return err
	}

	action := "added to"
	if args[1] == "delete" {
		action = "removed from"
	}
	return sendCrewReply(bot, update, fmt.Sprintf("@%s was %s crew %q", username, action, name))
}

func handleMentionCrew(bot *tgbot.BotAPI, update *tgbot.Update, name string) error {
	name = crews.NormalizeName(name)
	if err := crews.ValidateName(name); err != nil {
		return err
	}

	crew, err := getCrew(update, name)
	if err != nil {
		return err
	}
	if len(crew.Members) == 0 {
		return fmt.Errorf("Crew %q has no members", name)
	}

	mentions := make([]string, len(crew.Members))
	for i, member := range crew.Members {
		mentions[i] = "@" + member
	}
	return sendCrewReply(bot, update, strings.Join(mentions, " "))
}

func TryHandleCrewMention(bot *tgbot.BotAPI, update *tgbot.Update) (bool, error) {
	message := update.Message
	if message == nil || message.Command() != "" {
		return false, nil
	}

	mention, ok := singleCrewMention(message)
	if !ok {
		return false, nil
	}

	name := crews.NormalizeName(mention[1:])
	if err := crews.ValidateName(name); err != nil {
		return false, nil
	}

	crew, err := crews.GetCrewByChatAndName(message.Chat.ID, name)
	if err != nil {
		if errors.Is(err, crews.ErrCrewNotFound) {
			return false, nil
		}
		return true, err
	}

	if len(crew.Members) == 0 {
		return true, fmt.Errorf("Crew %q has no members", name)
	}
	mentions := make([]string, len(crew.Members))
	for i, member := range crew.Members {
		mentions[i] = "@" + member
	}

	return true, sendCrewReply(bot, update, strings.Join(mentions, " "))
}

func getCrew(update *tgbot.Update, name string) (*crews.Crew, error) {
	if err := crews.ValidateName(name); err != nil {
		return nil, err
	}

	crew, err := crews.GetCrewByChatAndName(update.Message.Chat.ID, name)
	if errors.Is(err, crews.ErrCrewNotFound) {
		return nil, fmt.Errorf("Crew %q doesn't exist in this chat", name)
	}
	return crew, err
}

func requireCrewManager(bot *tgbot.BotAPI, update *tgbot.Update, crew *crews.Crew) error {
	if update.Message.From == nil {
		return errors.New("A user account is required to manage a crew")
	}

	user, err := crews.UpsertUser(crews.UserDto{
		TelegramLogin: crews.NormalizeUsername(update.Message.From.UserName),
		FirstName:     update.Message.From.FirstName,
		LastName:      update.Message.From.LastName,
	})
	if err != nil {
		return err
	}
	if user.ID == crew.CreatorUserID {
		return nil
	}

	admin, err := isChatAdmin(bot, update)
	if err != nil {
		return err
	}
	if !admin {
		return errors.New("Only the crew creator or a chat administrator can manage this crew")
	}
	return nil
}

func parseCrewMembers(input string, creatorUsername string) ([]string, error) {
	if !validCrewMemberSeparators(input) {
		return nil, errors.New("Crew members must be separated by spaces, commas, or both")
	}

	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})

	members := []string{creatorUsername}
	seen := map[string]struct{}{creatorUsername: {}}
	additionalMembers := 0
	for _, part := range parts {
		member := crews.NormalizeUsername(part)
		if err := crews.ValidateUsername(member); err != nil {
			return nil, err
		}
		if member == creatorUsername {
			continue
		}
		if _, exists := seen[member]; exists {
			return nil, fmt.Errorf("Duplicate crew member @%s", member)
		}
		additionalMembers++
		if additionalMembers > 1 {
			return nil, errors.New("A crew can have only one additional member when it is created")
		}
		seen[member] = struct{}{}
		members = append(members, member)
	}
	if len(members) < 2 {
		return nil, errors.New("A crew must have at least 2 unique members, including its creator")
	}
	return members, nil
}

func validCrewMemberSeparators(input string) bool {
	runes := []rune(strings.TrimSpace(input))
	if len(runes) == 0 || runes[0] == ',' || runes[len(runes)-1] == ',' {
		return false
	}

	for i, r := range runes {
		if r != ',' {
			continue
		}

		before := i - 1
		for before >= 0 && unicode.IsSpace(runes[before]) {
			before--
		}
		after := i + 1
		for after < len(runes) && unicode.IsSpace(runes[after]) {
			after++
		}
		if before < 0 || after == len(runes) || runes[before] == ',' || runes[after] == ',' {
			return false
		}
	}
	return true
}

func singleCrewMention(message *tgbot.Message) (string, bool) {
	if strings.TrimSpace(message.Text) == "" || len(message.Entities) == 0 {
		return "", false
	}

	var mention string
	mentionCount := 0
	for _, entity := range message.Entities {
		if !entity.IsMention() {
			continue
		}
		value, ok := messageEntityText(message.Text, entity)
		if !ok {
			return "", false
		}
		mention = value
		mentionCount++
	}
	if mentionCount != 1 || strings.TrimSpace(message.Text) != mention || len(mention) < 2 || mention[0] != '@' {
		return "", false
	}
	return mention, true
}

func messageEntityText(text string, entity tgbot.MessageEntity) (string, bool) {
	encoded := utf16.Encode([]rune(text))
	start := entity.Offset
	end := entity.Offset + entity.Length
	if start < 0 || end < start || end > len(encoded) {
		return "", false
	}
	return string(utf16.Decode(encoded[start:end])), true
}

func sendCrewReply(bot *tgbot.BotAPI, update *tgbot.Update, text string) error {
	replyToMessageID := update.Message.MessageID
	if update.Message.ReplyToMessage != nil {
		replyToMessageID = update.Message.ReplyToMessage.MessageID
	}

	messageConfig := tgbot.NewMessage(update.Message.Chat.ID, text)
	messageConfig.ReplyToMessageID = replyToMessageID
	_, err := bot.Send(messageConfig)
	return err
}

func NewCrewCommandHandler() CrewCommandHandler {
	return CrewCommandHandler{CommandHandler: CommandHandler{command: "crew"}}
}
