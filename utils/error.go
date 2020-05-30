package utils

func CheckError(err error) bool {
	if err != nil {
		Log.Println(err)
		return true
	}
	return false
}

func CheckErrorExit(err error) bool {
	if err != nil {
		Log.Fatal(err)
		return true
	}
	return false
}