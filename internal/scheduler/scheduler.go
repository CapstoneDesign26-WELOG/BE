package scheduler

import (
	"log"
	"welog/internal/user"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron        *cron.Cron
	userService *user.UserService
}

func NewScheduler(userService *user.UserService) *Scheduler {
	c := cron.New()
	return &Scheduler{
		cron:        c,
		userService: userService,
	}
}

func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc("0 0 * * *", func() {
		log.Println("[Scheduler] 일일 토큰 리필 작업 시작...")

		if err := s.userService.RefillDailyTokens(); err != nil {
			log.Println("[Scheduler] 리필 중 에러 발생:", err)
		} else {
			log.Println("[Scheduler] 리필 성공!")
		}
	})

	if err != nil {
		log.Fatalf("스케줄러 등록 실패: %v", err)
	}

	s.cron.Start()
	log.Println("[Scheduler] 백그라운드 스케줄러가 시작되었습니다.")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("[Scheduler] 스케줄러가 종료되었습니다.")
}
