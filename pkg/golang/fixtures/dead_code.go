package fixtures

import "fmt"

func main() {
	fmt.Println(ValidateUser("John", "", 25))
	fmt.Println(ValidateUser("", "Doe", 17))
}

func ValidateUser(firstName, lastName string, age int) string {
	if firstName == "" {
		return "Error: First name is required"
	}
	if lastName == "" {
		return "Error: Last name is required"
	}
	if age < 18 {
		return "Error: User must be 18 or older"
	}
	return "Valid user"
}
