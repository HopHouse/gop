package gopOsintEmailGen

import "github.com/hophouse/gop/utils/logger"

// RunEmailGen will create all the variations of email based on the inputed data.
func RunEmailGen(firstname string, surname string, domain string, delimiters []string) {
	// [firstname][surname]@[domain]
	logger.Printf("%s%s@%s\n", firstname, surname, domain)

	if len(firstname) > 1 {
		// [f][surname]@[domain]
		logger.Printf("%c%s@%s\n", firstname[0], surname, domain)
	}

	if len(firstname) > 2 {
		// [fi][surname]@[domain]
		logger.Printf("%s%s@%s\n", firstname[0:2], surname, domain)
	}

	if len(firstname) > 3 {
		// [fir][surname]@[domain]
		logger.Printf("%s%s@%s\n", firstname[0:3], surname, domain)
	}

	if len(firstname) > 4 {
		// [firs][surname]@[domain]
		logger.Printf("%s%s@%s\n", firstname[0:4], surname, domain)
	}

	// [surname][firstname]@[domain]
	logger.Printf("%s%s@%s\n", surname, firstname, domain)

	if len(firstname) > 1 {
		// [surname][f]@[domain]
		logger.Printf("%s%c@%s\n", surname, firstname[0], domain)
	}

	if len(firstname) > 2 {
		// [surname][fi]@[domain]
		logger.Printf("%s%s@%s\n", surname, firstname[0:2], domain)
	}

	if len(firstname) > 3 {
		// [surname][fir]@[domain]
		logger.Printf("%s%s@%s\n", surname, firstname[0:3], domain)
	}

	if len(firstname) > 4 {
		// [surname][firs]@[domain]
		logger.Printf("%s%s@%s\n", surname, firstname[0:4], domain)
	}

	for _, delimiter := range delimiters {
		// [firstname][delimiters][surname]@[domain]
		logger.Printf("%s%s%s@%s\n", firstname, delimiter, surname, domain)

		if len(firstname) > 1 {
			// [f][delimiters][surname]@[domain]
			logger.Printf("%c%s%s@%s\n", firstname[0], delimiter, surname, domain)
		}

		if len(firstname) > 2 {
			// [fi][delimiters][surname]@[domain]
			logger.Printf("%s%s%s@%s\n", firstname[0:2], delimiter, surname, domain)
		}

		if len(firstname) > 3 {
			// [fir][delimiters][surname]@[domain]
			logger.Printf("%s%s%s@%s\n", firstname[0:3], delimiter, surname, domain)
		}

		if len(firstname) > 4 {
			// [firs][delimiters][surname]@[domain]
			logger.Printf("%s%s%s@%s\n", firstname[0:4], delimiter, surname, domain)
		}

		// [surname][delimiters][firstname]@[domain]
		logger.Printf("%s%s%s@%s\n", surname, delimiter, firstname, domain)

		if len(firstname) > 1 {
			// [surname][delimiters][f]@[domain]
			logger.Printf("%s%s%c@%s\n", surname, delimiter, firstname[0], domain)
		}

		if len(firstname) > 2 {
			// [surname][delimiters][fi]@[domain]
			logger.Printf("%s%s%s@%s\n", surname, delimiter, firstname[0:2], domain)
		}

		if len(firstname) > 3 {
			// [surname][delimiters][fir]@[domain]
			logger.Printf("%s%s%s@%s\n", surname, delimiter, firstname[0:3], domain)
		}

		if len(firstname) > 4 {
			// [surname][delimiters][firs]@[domain]
			logger.Printf("%s%s%s@%s\n", surname, delimiter, firstname[0:4], domain)
		}
	}
}
