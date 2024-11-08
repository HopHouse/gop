package utils

import (
	"sync"

	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var (
	ScreenshotBar *ProgessBar
	CrawlerBar    *ProgessBar
)

type ProgessBar struct {
	waitGroup sync.WaitGroup
	mpbBar    *mpb.Bar
	name      string
	total     int
	mutex     *sync.Mutex
}

type WaitGroupBar struct {
	waitGroup sync.WaitGroup
	progress  *mpb.Progress
	bars      []*ProgessBar
}

func InitWaitGroupBar() *WaitGroupBar {
	var groupBar WaitGroupBar
	groupBar.progress = mpb.New(mpb.WithWidth(1))
	return &groupBar
}

func NewBar(name string) *ProgessBar {
	newBar := ProgessBar{}
	newBar.name = name
	newBar.total = 0
	newBar.mutex = &sync.Mutex{}
	newBar.waitGroup = sync.WaitGroup{}
	newBar.mpbBar = mpb.New(mpb.WithWidth(1)).AddBar(1,
		// mpb.NewSpinnerFiller([]string{}, mpb.SpinnerOnLeft),
		mpb.PrependDecorators(decor.Name("[")),
		mpb.AppendDecorators(
			decor.Name("] ["),
			decor.Name(name),
			decor.Name("] ["),
			decor.Counters(0, "%d / %d"),
			decor.OnComplete(decor.Name("] [Running]"), "] [Finished]"),
		),
	)

	return &newBar
}

func (groupBar *WaitGroupBar) AddBar(name string, main bool) *ProgessBar {
	newBar := NewBar(name)

	newBar.waitGroup = groupBar.waitGroup
	groupBar.bars = append(groupBar.bars, newBar)

	return newBar
}

func (groupBar *WaitGroupBar) Wait() {
	groupBar.waitGroup.Wait()
	for _, item := range groupBar.bars {
		item.Wait()
		item.mpbBar.SetTotal(int64(item.total), true)
	}
	groupBar.progress.Wait()
}

func (bar *ProgessBar) Add(delta int) {
	bar.mutex.Lock()
	bar.total += delta
	bar.waitGroup.Add(delta)
	bar.mutex.Unlock()
}

func (bar *ProgessBar) AddAndIncrementTotal(delta int) {
	bar.mutex.Lock()
	bar.total += delta
	bar.mpbBar.SetTotal(int64(bar.total), false)
	bar.waitGroup.Add(delta)
	bar.mutex.Unlock()
}

func (bar *ProgessBar) Done() {
	bar.mutex.Lock()
	bar.waitGroup.Done()
	bar.mpbBar.IncrBy(1)
	bar.mutex.Unlock()
}

func (bar *ProgessBar) Wait() {
	bar.waitGroup.Wait()
	bar.SetTotal(bar.total)
	bar.mpbBar.SetTotal(int64(bar.total), true)
}

func (bar *ProgessBar) SetTotal(total int) {
	bar.mutex.Lock()
	bar.total = total
	bar.mpbBar.SetTotal(int64(bar.total), false)
	bar.mutex.Unlock()
}
