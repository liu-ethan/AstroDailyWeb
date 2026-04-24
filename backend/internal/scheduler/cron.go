package scheduler

import (
	"context"
	"log/slog"

	"github.com/robfig/cron/v3"

	"astrodailyweb/backend/internal/service"
)

type Scheduler struct {
	cron   *cron.Cron
	logger *slog.Logger
	user   service.UserService
	f      service.FortuneService
}

// New 创建定时任务调度器。
// 参数：logger - 日志实例；userSvc - 用户服务；fortuneSvc - 运势服务。
// 返回：*Scheduler - 调度器实例。
func New(logger *slog.Logger, userSvc service.UserService, fortuneSvc service.FortuneService) *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		logger: logger,
		user:   userSvc,
		f:      fortuneSvc,
	}
}

// RegisterJobs 注册定时任务。
// 参数：无。
// 返回：error - 注册失败时返回错误。
func (s *Scheduler) RegisterJobs() error {
	// 每天 07:30:00 预生成订阅用户的运势并执行历史清理。
	_, err := s.cron.AddFunc("0 30 7 * * *", func() {
		ctx := context.Background()
		users, listErr := s.user.ListSubscribedUsers(ctx)
		if listErr != nil {
			s.logger.Error("list subscribed users failed", "err", listErr)
			return
		}
		if genErr := s.f.GenerateForSubscribedUsers(ctx, users); genErr != nil {
			s.logger.Error("generate scheduled fortunes failed", "err", genErr)
		}
		if cleanErr := s.f.CleanupHistory(ctx, 7); cleanErr != nil {
			s.logger.Error("cleanup history failed", "err", cleanErr)
		}
	})
	return err
}

// Start 启动调度器。
// 参数：无。
// 返回：无。
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop 停止调度器并等待任务退出。
// 参数：无。
// 返回：无。
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
}
