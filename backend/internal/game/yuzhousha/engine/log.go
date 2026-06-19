package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	logMu     sync.Mutex
	logFile   *os.File
	logInited bool
)

const logDir = "logs"

func initLogFile() {
	if logInited {
		return
	}
	logInited = true
	// 查找项目根目录（向上找有 go.mod 的目录）
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		if _, e := os.Stat(filepath.Join(dir, "go.mod")); e == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
	logPath := filepath.Join(dir, logDir)
	os.MkdirAll(logPath, 0755)
	filename := filepath.Join(logPath, fmt.Sprintf("game_%s.log", time.Now().Format("20060102_150405")))
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	logFile = f
	fmt.Fprintf(os.Stderr, "[LOG] logging to %s\n", filename)
}

func Logf(format string, args ...interface{}) {
	logMu.Lock()
	defer logMu.Unlock()
	initLogFile()
	if logFile == nil {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	shortFile := filepath.Base(file)
	ts := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logFile, "[%s] %s:%d %s\n", ts, shortFile, line, msg)
	logFile.Sync()
}

func LogGameState(g *Game, tag string) {
	if g == nil {
		Logf("%s: game is nil", tag)
		return
	}
	Logf("%s: Phase=%s TurnStep=%s CurrentTurn=%d", tag, g.Phase, g.TurnStep, g.CurrentTurn)
	if g.Pending != nil {
		p := g.Pending
		Logf("%s Pending: ResponseMode=%s ActorSeat=%d TargetIndex=%d SourceIndex=%d EffectTarget=%d AllowWuxiek=%v TaoYuanQueue=%v WuguPickSeat=%d Queue=%v QueueIdx=%d ChainLen=%d Card=%s",
			tag, p.ResponseMode, p.ActorSeat, p.TargetIndex, p.SourceIndex, p.EffectTarget,
			p.AllowWuxiek, p.TaoYuanQueue, p.WuguPickSeat, p.ResponseQueue, p.ResponseIndex, len(p.WuxiekChain), p.Card.Kind)
	} else {
		Logf("%s Pending: nil", tag)
	}
	for i, pl := range g.Players {
		Logf("%s Player[%d]=%s HP=%d/%d Hand=%d AI=%v", tag, i, pl.Name, pl.HP, pl.MaxHP, len(pl.Hand), pl.IsAI)
	}
}
