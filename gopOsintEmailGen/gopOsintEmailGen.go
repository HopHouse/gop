package gopOsintEmailGen

import "fmt"

// RunEmailGen will create all the variations of email based on the inputed data.
func RunEmailGen(firstname string, surname string, domain string, delimiters []string) {
	for _, delimiter := range delimiters {
		// [firstname][delimiters][surname]@[domain]
		fmt.Printf("%s%s%s@%s\n", firstname, delimiter, surname, domain)

		if len(firstname) > 1 {
			// [f][delimiters][surname]@[domain]
			fmt.Printf("%c%s%s@%s\n", firstname[0], delimiter, surname, domain)
		}

		if len(firstname) > 2 {
			// [fi][delimiters][surname]@[domain]
			fmt.Printf("%s%s%s@%s\n", firstname[0:1], delimiter, surname, domain)
		}

		if len(firstname) > 3 {
			// [fir][delimiters][surname]@[domain]
			fmt.Printf("%s%s%s@%s\n", firstname[0:2], delimiter, surname, domain)
		}

		if len(firstname) > 4 {
			// [firs][delimiters][surname]@[domain]
			fmt.Printf("%s%s%s@%s\n", firstname[0:3], delimiter, surname, domain)
		}

		// [surname][delimiters][firstname]@[domain]
		fmt.Printf("%s%s%s@%s\n", surname, delimiter, firstname, domain)

		if len(firstname) > 1 {
			// [surname][delimiters][f]@[domain]
			fmt.Printf("%s%s%c@%s\n", surname, delimiter, firstname[0], domain)
		}

		if len(firstname) > 2 {
			// [surname][delimiters][fi]@[domain]
			fmt.Printf("%s%s%s@%s\n", surname, delimiter, firstname[0:1], domain)
		}

		if len(firstname) > 3 {
			// [surname][delimiters][fir]@[domain]
			fmt.Printf("%s%s%s@%s\n", surname, delimiter, firstname[0:2], domain)
		}

		if len(firstname) > 4 {
			// [surname][delimiters][firs]@[domain]
			fmt.Printf("%s%s%s@%s\n", surname, delimiter, firstname[0:3], domain)
		}
	}
}
