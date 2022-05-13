package gopOsintUsernameGen

import "fmt"

// RunEmailGen will create all the variations of email based on the inputed data.
func RunUsernameGen(firstname string, surname string, delimiters []string) {
	// [firstname][surname]
	fmt.Printf("%s%s\n", firstname, surname)

	if len(firstname) > 1 {
		// [f][surname]
		fmt.Printf("%c%s\n", firstname[0], surname)
	}

	if len(firstname) > 2 {
		// [fi][surname]
		fmt.Printf("%s%s\n", firstname[0:2], surname)
	}

	if len(firstname) > 3 {
		// [fir][surname]
		fmt.Printf("%s%s\n", firstname[0:3], surname)
	}

	if len(firstname) > 4 {
		// [firs][surname]
		fmt.Printf("%s%s\n", firstname[0:4], surname)
	}

	// [surname][firstname]
	fmt.Printf("%s%s\n", surname, firstname)

	if len(firstname) > 1 {
		// [surname][f]
		fmt.Printf("%s%c\n", surname, firstname[0])
	}

	if len(firstname) > 2 {
		// [surname][fi]
		fmt.Printf("%s%s\n", surname, firstname[0:2])
	}

	if len(firstname) > 3 {
		// [surname][fir]
		fmt.Printf("%s%s\n", surname, firstname[0:3])
	}

	if len(firstname) > 4 {
		// [surname][firs]
		fmt.Printf("%s%s\n", surname, firstname[0:4])
	}

	for _, delimiter := range delimiters {
		// [firstname][delimiters][surname]
		fmt.Printf("%s%s%s\n", firstname, delimiter, surname)

		if len(firstname) > 1 {
			// [f][delimiters][surname]
			fmt.Printf("%c%s%s\n", firstname[0], delimiter, surname)
		}

		if len(firstname) > 2 {
			// [fi][delimiters][surname]
			fmt.Printf("%s%s%s\n", firstname[0:2], delimiter, surname)
		}

		if len(firstname) > 3 {
			// [fir][delimiters][surname]
			fmt.Printf("%s%s%s\n", firstname[0:3], delimiter, surname)
		}

		if len(firstname) > 4 {
			// [firs][delimiters][surname]
			fmt.Printf("%s%s%s\n", firstname[0:4], delimiter, surname)
		}

		// [surname][delimiters][firstname]
		fmt.Printf("%s%s%s\n", surname, delimiter, firstname)

		if len(firstname) > 1 {
			// [surname][delimiters][f]
			fmt.Printf("%s%s%c\n", surname, delimiter, firstname[0])
		}

		if len(firstname) > 2 {
			// [surname][delimiters][fi]
			fmt.Printf("%s%s%s\n", surname, delimiter, firstname[0:2])
		}

		if len(firstname) > 3 {
			// [surname][delimiters][fir]
			fmt.Printf("%s%s%s\n", surname, delimiter, firstname[0:3])
		}

		if len(firstname) > 4 {
			// [surname][delimiters][firs]
			fmt.Printf("%s%s%s\n", surname, delimiter, firstname[0:4])
		}
	}
}
