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
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/hophouse/gop/utils/logger"
	"github.com/spf13/cobra"
)

var (
	secondOption     int
	minuteOption     int
	hourOption       int
	dayOption        int
	monthOption      int
	yearOption       int
	executionTime    time.Time
	cmdOption        string
	plusSecondOption int
	plusMinuteOption int
	plusHourOption   int
	plusDayOption    int
)

// hostCmd represents the host command
var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Schedule a command to be executed at a precise time. If an option is not defined, the value of the current date will be taken.",
	Long:  "Schedule a command to be executed at a precise time. If an option is not defined, the value of the current date will be taken.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.NewLoggerStdout()
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		// Get time
		t := time.Now()

		if secondOption == -1 {
			secondOption = t.Second()
		}

		if minuteOption == -1 {
			minuteOption = t.Minute()
		}

		if hourOption == -1 {
			hourOption = t.Hour()
		}

		if dayOption == -1 {
			dayOption = t.Day()
		}

		if monthOption == -1 {
			monthOption = int(t.Month())
		}

		if yearOption == -1 {
			yearOption = t.Year()
		}

		executionTime = time.Date(yearOption, ((time.Month)(monthOption)), dayOption, hourOption, minuteOption, secondOption, 0, t.Location())

		executionTime = executionTime.Add(time.Second * time.Duration(plusSecondOption))
		executionTime = executionTime.Add(time.Minute * time.Duration(plusMinuteOption))
		executionTime = executionTime.Add(time.Hour * time.Duration(plusHourOption))
		executionTime = executionTime.Add(time.Hour * 24 * time.Duration(plusDayOption))
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger.Printf("[+] Secheduled execution time : %s\n", executionTime.String())
		if cmdOption == "" {
			erroMsg := "No command passed as parameter. Please give a command as argument.\n"
			log.Fatal(erroMsg)
		}
		logger.Printf("[+] Scheduled command : %s\n", cmdOption)

		// Time is already passed, then execute the command.
		if executionTime.Before(time.Now()) {
			executeCommand(cmdOption)
		} else {
			// Time is in the future, then sleep until time is reached and execute the command
			waitDuration := time.Until(executionTime)
			logger.Printf("[+] Command will be executed in %s.\n", waitDuration)

			logger.Printf("\n")
			time.Sleep(waitDuration)

			executeCommand(cmdOption)
		}
	},
}

func executeCommand(command string) {
	var cmd *exec.Cmd
	args := strings.Fields(command)

	logger.Printf("$> %s\n", command)
	if len(args) == 1 {
		cmd = exec.Command(args[0])
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("[!] Error : %s\n", err)
		log.Fatal(errMsg)
	}

	// Display the output
	logger.Printf("%s\n", stdoutStderr)
}

func init() {
	scheduleCmd.PersistentFlags().IntVarP(&secondOption, "second", "s", -1, "Second of the minute.")
	scheduleCmd.PersistentFlags().IntVarP(&minuteOption, "minute", "m", -1, "Minute of the hour.")
	scheduleCmd.PersistentFlags().IntVarP(&hourOption, "hour", "", -1, "Hour of the day.")
	scheduleCmd.PersistentFlags().IntVarP(&dayOption, "day", "", -1, "Days of the month.")
	scheduleCmd.PersistentFlags().IntVarP(&monthOption, "month", "", -1, "Month of the year.")
	scheduleCmd.PersistentFlags().IntVarP(&yearOption, "year", "", -1, "Year were the command will be launched.")

	scheduleCmd.PersistentFlags().StringVarP(&cmdOption, "command", "c", "", "Command to execute.")

	scheduleCmd.PersistentFlags().IntVarP(&plusSecondOption, "plus-seconds", "", 0, "Add this number of seconds from the execution to execute the command.")
	scheduleCmd.PersistentFlags().IntVarP(&plusMinuteOption, "plus-minutes", "", 0, "Add this number of minutes from the execution to execute the command.")
	scheduleCmd.PersistentFlags().IntVarP(&plusHourOption, "plus-hours", "", 0, "Add this number of hours from the execution to execute the command.")
	scheduleCmd.PersistentFlags().IntVarP(&plusDayOption, "plus-days", "", 0, "Add this number of days from the execution to execute the command.")
}
