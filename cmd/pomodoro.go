/*
Copyright © 2020 Hophouse <contact@hophouse.fr>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"strings"
	"time"

	"github.com/hophouse/gop/notification"
	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var (
	periodOption     int
	shortBreakOption int
	longBreakOption  int
	cycleNbOption    int
)

// pomodoroCmd represents the host command
var pomodoroCmd = &cobra.Command{
	Use:  "pomodoro",
	Long: "Resolve hostname to get the IP address.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerStdout()
	},
	Run: func(cmd *cobra.Command, args []string) {
		for {
			for i := 1; i < cycleNbOption+1; i++ {
				displayBar("Work", periodOption)

				notification.NotifyAndWait("Short Break")
				displayBar("Short break", shortBreakOption)
				notification.NotifyAndWait("Work")
			}

			displayBar("Work", periodOption)

			notification.NotifyAndWait("Long break")
			displayBar("Long break", longBreakOption)
			notification.NotifyAndWait("Work")
		}
	},
}

func init() {
	rootCmd.AddCommand(pomodoroCmd)

	pomodoroCmd.PersistentFlags().IntVarP(&periodOption, "period", "p", 25, "Time in minutes allowed to work.")
	pomodoroCmd.PersistentFlags().IntVarP(&shortBreakOption, "short", "s", 5, "Time in minutes allowed to the short break.")
	pomodoroCmd.PersistentFlags().IntVarP(&longBreakOption, "long", "l", 15, "Time in minutes allowed to the long break.")
	pomodoroCmd.PersistentFlags().IntVarP(&cycleNbOption, "cycle", "c", 3, "Number of short break before a long break happen.")
}

func displayBar(name string, duration int) {
	p := mpb.New(mpb.WithWidth(64))

	total := 100

	// adding a single bar, which will inherit container's width
	bar := p.Add(int64(total),
		// progress bar filler with customized style
		mpb.NewBarFiller(mpb.BarStyle().Lbound("╢").Filler("▌").Tip("▌").Padding("░").Rbound("╟")),
		mpb.PrependDecorators(
			decor.Name(computeString(name)),
			decor.Elapsed(decor.ET_STYLE_HHMMSS),
		),
		mpb.AppendDecorators(decor.Percentage(decor.WCSyncSpace)),
	)

	for i := 0; i < total; i++ {
		time.Sleep(time.Duration(duration) * time.Duration(time.Minute) / time.Duration(total))
		bar.Increment()
	}

	// wait for our bar to complete and flush
	p.Wait()
}

func computeString(name string) string {
	total := 15
	spacesNb := total - len(name)
	if spacesNb < 0 {
		return name
	}

	spaces := strings.Repeat(" ", spacesNb)

	return string(name + spaces)
}
