package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand/v2" //これmath/randより効率いい。
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/kou-etal/go_todo_app/internal/config"
	taskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
	txrunner "github.com/kou-etal/go_todo_app/internal/infra/db/tx"
	taskeventrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/event"
	taskrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/task"
	userrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/user"
	"github.com/kou-etal/go_todo_app/internal/infra/security"
)

const maxSeedUsers = 10000

var taskTitles = []string{
	"Buy groceries",
	"Write report",
	"Fix login bug",
	"Deploy staging",
	"Review PR",
	"Update docs",
	"Plan sprint",
	"Setup CI/CD",
	"Refactor auth",
	"Add unit tests",
}

var taskDescs = []string{
	"This task needs to be completed as soon as possible.",
	"Please review the requirements before starting.",
	"Check the documentation for implementation details.",
	"Coordinate with the team before making changes.",
	"Run all tests after completing this task.",
}

var dueOptions = []dtask.DueOption{
	dtask.Due7Days,
	dtask.Due14Days,
	dtask.Due21Days,
	dtask.Due30Days,
}

// seedDeps は seed 専用の Tx 内 DI。
// 既存の usecase/tx/deps.go は変更せず、ここにローカル定義する。
type seedDeps struct {
	userRepo      duser.UserRepository
	taskRepo      dtask.TaskRepository
	taskEventRepo taskevent.TaskEventRepository
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("seed terminated with error: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	//TODO:遅すぎたらgoroutine使う。
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	//flagはGoの標準ライブラリで、コマンドライン引数をパースするためのパッケージ
	//返り値はポインタ。flag.Parse()まではdefault値。flag.Parse()でCLI。
	//name/default/explain
	users := flag.Int("users", 50, "number of seed users to create")
	tasksPerUser := flag.Int("tasks-per-user", 30, "number of tasks per user")
	prefix := flag.String("prefix", "seed", "prefix for email/username")
	clean := flag.Bool("clean", false, "delete existing seed data before inserting")
	flag.Parse()

	//flag バリデーション
	if *users <= 0 {
		return fmt.Errorf("--users must be positive, got %d", *users)
	}
	if *users > maxSeedUsers {
		return fmt.Errorf("too many users: %d (max %d)", *users, maxSeedUsers)
	}
	if *tasksPerUser <= 0 {
		return fmt.Errorf("--tasks-per-user must be positive, got %d", *tasksPerUser)
	}

	// userName の 20 文字制約チェック
	digits := max(3, len(fmt.Sprintf("%d", *users)))
	//fmt.Sprintf("%s-user-%0*d", "seed", 3, 50)でseed-user-050。maxをチェックする。
	sampleName := fmt.Sprintf("%s-user-%0*d", *prefix, digits, *users)
	if len(sampleName) > 20 {
		return fmt.Errorf("prefix %q is too long: user_name %q exceeds 20 chars", *prefix, sampleName)
	}

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	xdb, closeDB, err := db.NewMySQL(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer closeDB()

	// 既存 seed 存在チェック
	pattern := *prefix + "-%@example.com" //Goこれできる。ワイルドカードのための%
	//seed-で始まって@example.comで終わる
	existing, err := countSeedUsers(ctx, xdb, pattern)
	if err != nil {
		return fmt.Errorf("count seed users: %w", err)
	}

	if *clean {
		if existing > 0 {
			// seed は使い捨てバッチ CLI で、JSON 構造ログいらない->slog使わない。
			log.Printf("seed: cleaning %d existing seed users (prefix=%s)", existing, *prefix)
			if err := deleteSeedUsers(ctx, xdb, pattern); err != nil {
				return fmt.Errorf("clean seed data: %w", err)
			}
		}
	} else if existing > 0 { //これはif cleanに対するelse if。clean必須やのにcleanしていない。
		return fmt.Errorf("seed users already exist (%d). use --clean to reset", existing)
	}

	// bcrypt 1 回
	hasher := security.NewBcryptHasher(12)
	passVO, err := duser.NewUserPasswordFromPlain(cfg.SeedPassword, hasher) //domainバリデーションを通す。
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}
	cachedHash := passVO.Hash()

	// TxRunner
	txOpts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	makeDeps := func(q db.QueryerExecer) seedDeps {
		return seedDeps{
			userRepo:      userrepo.NewRepository(q),
			taskRepo:      taskrepo.New(q),
			taskEventRepo: taskeventrepo.New(q),
		}
	}
	runner := txrunner.New[seedDeps](xdb, txOpts, makeDeps)

	//seed
	totalTasks := *users * *tasksPerUser
	log.Printf("seed: starting (%d users, %d tasks/user, total %d tasks)", *users, *tasksPerUser, totalTasks)
	start := time.Now()
	for i := 1; i <= *users; i++ {
		select {
		case <-ctx.Done():
			log.Printf("seed: interrupted at %d/%d", i, *users)
			return ctx.Err()
		default:
		}

		now := time.Now()
		num := fmt.Sprintf("%0*d", digits, i)
		emailStr := fmt.Sprintf("%s-%s@example.com", *prefix, num)
		nameStr := fmt.Sprintf("%s-user-%s", *prefix, num)

		u, err := buildUser(emailStr, nameStr, cachedHash, now)
		if err != nil {
			return fmt.Errorf("build user %s: %w", emailStr, err)
		}

		tasks, events, err := buildTasks(u, *tasksPerUser, now)
		if err != nil {
			return fmt.Errorf("build tasks for user %s: %w", emailStr, err)
		}

		if err := runner.WithinTx(ctx, func(ctx context.Context, deps seedDeps) error {
			if err := deps.userRepo.Store(ctx, u); err != nil {
				return fmt.Errorf("store user: %w", err)
			}
			for j := range tasks {
				if err := deps.taskRepo.Store(ctx, tasks[j]); err != nil {
					return fmt.Errorf("store task %d: %w", j+1, err)
				}
				if err := deps.taskEventRepo.Insert(ctx, events[j]); err != nil {
					return fmt.Errorf("insert event %d: %w", j+1, err)
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("seed user %d/%d (%s): %w", i, *users, emailStr, err)
		}

		log.Printf("seed: [%d/%d] user=%s tasks=%d", i, *users, emailStr, *tasksPerUser)
	}

	elapsed := time.Since(start).Round(time.Millisecond)
	log.Printf("seed: completed in %s (%d users, %d tasks, %d events)", elapsed, *users, totalTasks, totalTasks)
	return nil
}

func buildUser(email, name, cachedHash string, now time.Time) (*duser.User, error) {
	emailVO, err := duser.NewUserEmail(email)
	if err != nil {
		return nil, fmt.Errorf("email: %w", err)
	}
	nameVO, err := duser.NewUserName(name)
	if err != nil {
		return nil, fmt.Errorf("name: %w", err)
	}
	passVO, err := duser.ReconstructUserPassword(cachedHash)
	if err != nil {
		return nil, fmt.Errorf("password: %w", err)
	}

	u := duser.NewUser(emailVO, passVO, nameVO, now)
	u.VerifyEmail(now)
	return u, nil
}

func buildTasks(u *duser.User, count int, now time.Time) ([]*dtask.Task, []*taskevent.TaskEvent, error) {
	tasks := make([]*dtask.Task, 0, count)
	events := make([]*taskevent.TaskEvent, 0, count)

	for j := 0; j < count; j++ {
		title, err := dtask.NewTaskTitle(taskTitles[rand.IntN(len(taskTitles))])
		if err != nil {
			return nil, nil, fmt.Errorf("title: %w", err)
		}
		desc, err := dtask.NewTaskDescription(taskDescs[rand.IntN(len(taskDescs))])
		if err != nil {
			return nil, nil, fmt.Errorf("description: %w", err)
		}
		opt := dueOptions[rand.IntN(len(dueOptions))]
		due, err := dtask.NewDueDateFromOption(now, opt)
		if err != nil {
			return nil, nil, fmt.Errorf("due date: %w", err)
		}

		t := dtask.NewTask(u.ID(), title, desc, due, now)
		tasks = append(tasks, t)

		reqID := taskevent.RequestID(uuid.NewString())
		ev := taskevent.NewCreatedEvent(u.ID(), t.ID(), reqID, now, taskevent.CreatedPayload{})
		events = append(events, ev)
	}

	return tasks, events, nil
}

func countSeedUsers(ctx context.Context, q db.QueryerExecer, pattern string) (int, error) {
	var count int
	err := q.GetContext(ctx, &count, "SELECT COUNT(*) FROM users WHERE email LIKE ?", pattern)
	return count, err
}

func deleteSeedUsers(ctx context.Context, q db.QueryerExecer, pattern string) error {
	_, err := q.ExecContext(ctx, "DELETE FROM users WHERE email LIKE ?", pattern)
	return err
}
