package runner

// wrap check if err statement in function for more readable code...
func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
