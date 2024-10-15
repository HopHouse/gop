package gopOsintUsernameGen

import "github.com/hophouse/gop/utils/logger"

// RunEmailGen will create all the variations of email based on the inputed data.
func RunUsernameGen(firstname string, surname string, delimiters []string) {
	// [firstname][surname]
	logger.Printf("%s%s\n", firstname, surname)

	if len(firstname) > 1 {
		// [f][surname]
		logger.Printf("%c%s\n", firstname[0], surname)
	}

	if len(firstname) > 2 {
		// [fi][surname]
		logger.Printf("%s%s\n", firstname[0:2], surname)
	}

	if len(firstname) > 3 {
		// [fir][surname]
		logger.Printf("%s%s\n", firstname[0:3], surname)
	}

	if len(firstname) > 4 {
		// [firs][surname]
		logger.Printf("%s%s\n", firstname[0:4], surname)
	}

	// [surname][firstname]
	logger.Printf("%s%s\n", surname, firstname)

	if len(firstname) > 1 {
		// [surname][f]
		logger.Printf("%s%c\n", surname, firstname[0])
	}

	if len(firstname) > 2 {
		// [surname][fi]
		logger.Printf("%s%s\n", surname, firstname[0:2])
	}

	if len(firstname) > 3 {
		// [surname][fir]
		logger.Printf("%s%s\n", surname, firstname[0:3])
	}

	if len(firstname) > 4 {
		// [surname][firs]
		logger.Printf("%s%s\n", surname, firstname[0:4])
	}

	for _, delimiter := range delimiters {
		// [firstname][delimiters][surname]
		logger.Printf("%s%s%s\n", firstname, delimiter, surname)

		if len(firstname) > 1 {
			// [f][delimiters][surname]
			logger.Printf("%c%s%s\n", firstname[0], delimiter, surname)
		}

		if len(firstname) > 2 {
			// [fi][delimiters][surname]
			logger.Printf("%s%s%s\n", firstname[0:2], delimiter, surname)
		}

		if len(firstname) > 3 {
			// [fir][delimiters][surname]
			logger.Printf("%s%s%s\n", firstname[0:3], delimiter, surname)
		}

		if len(firstname) > 4 {
			// [firs][delimiters][surname]
			logger.Printf("%s%s%s\n", firstname[0:4], delimiter, surname)
		}

		// [surname][delimiters][firstname]
		logger.Printf("%s%s%s\n", surname, delimiter, firstname)

		if len(firstname) > 1 {
			// [surname][delimiters][f]
			logger.Printf("%s%s%c\n", surname, delimiter, firstname[0])
		}

		if len(firstname) > 2 {
			// [surname][delimiters][fi]
			logger.Printf("%s%s%s\n", surname, delimiter, firstname[0:2])
		}

		if len(firstname) > 3 {
			// [surname][delimiters][fir]
			logger.Printf("%s%s%s\n", surname, delimiter, firstname[0:3])
		}

		if len(firstname) > 4 {
			// [surname][delimiters][firs]
			logger.Printf("%s%s%s\n", surname, delimiter, firstname[0:4])
		}
	}
}
