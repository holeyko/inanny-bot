package poll

import (
	"log"
	"sync"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	bot     *tgbot.BotAPI
	cron    *cron.Cron
	entries map[int64]cron.EntryID
	mu      sync.Mutex
}

var scheduler *Scheduler

func StartScheduler(bot *tgbot.BotAPI) error {
	scheduler = &Scheduler{
		bot:     bot,
		cron:    cron.New(),
		entries: map[int64]cron.EntryID{},
	}

	polls, err := GetCronPolls()
	if err == nil {
		for _, poll := range polls {
			if err := scheduler.Register(poll); err != nil {
				log.Println("Error while registering cron poll", poll.ID, err)
			}
		}
	}

	scheduler.cron.Start()
	return err
}

func RegisterCronPoll(poll *StoredPoll) error {
	if scheduler == nil {
		return nil
	}
	return scheduler.Register(poll)
}

func RemoveCronPoll(id int64) {
	if scheduler == nil {
		return
	}
	scheduler.Remove(id)
}

func (s *Scheduler) Register(poll *StoredPoll) error {
	if err := ValidateCronExpr(poll.CronExpr); err != nil {
		return err
	}

	pollCopy := *poll
	entryID, err := s.cron.AddFunc(poll.CronExpr, func() {
		if err := SendPollToChat(s.bot, &pollCopy.Poll, pollCopy.ChatID); err != nil {
			log.Println("Error while sending cron poll", pollCopy.ID, err)
		}
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.entries[poll.ID] = entryID
	s.mu.Unlock()

	return nil
}

func (s *Scheduler) Remove(id int64) {
	s.mu.Lock()
	entryID, ok := s.entries[id]
	if ok {
		delete(s.entries, id)
	}
	s.mu.Unlock()

	if ok {
		s.cron.Remove(entryID)
	}
}
