package main

const foo = `PT1H4M`

// type parser struct {
// 	sql   string      // The query to parse
// 	i     int         // Where we are in the query
// 	query query.Query // The "query struct" we'll build
// 	step  step        // What's this? Read on...
// }

// const logstring string = `# Line1
// # Line2
// Continued line2
// Continued line2
// # line3`

// func crunchSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {

// 	if atEOF && len(data) == 0 {
// 		return 0, nil, nil
// 	}

// 	if i := strings.IndexRune(string(data), IsLetter(); i >= 0 {
// 		return i + 1, data[0:i], nil
// 	}

// 	if atEOF {
// 		return len(data), data, nil
// 	}

// 	return
// }

func main() {
	// buf := ""
	// for pos, char := range foo {
	// 	if char == unicode.IsLetter(char) {

	// 	}
	// 	fmt.Printf("character %c at pos %d.\n", char, pos)
	// }
	// r := regexp.MustCompile(`(\d+\w)`)
	// fmt.Println(r.FindAllStringSubmatch(foo, -1))
	// parts := []string{}
	// part := []byte{}
	// for _, char := range foo {
	// 	if unicode.IsLetter(char) {
	// 		fmt.Println(string(char))
	// 	}
	// }
	// scanner := bufio.NewScanner(strings.NewReader(logstring))
	// scanner.Split(crunchSplitFunc)

	// for scanner.Scan() {
	// 	log.Print(scanner.Text())
	// } // lastQuote := rune(0)
	// f := func(c rune) bool {
	// 	fmt.Println(string(c))
	// 	switch {
	// 	case unicode.IsLetter(c):
	// 		return true
	// 	// case c == lastQuote:
	// 	// 	lastQuote = rune(0)
	// 	// 	return false
	// 	// case lastQuote != rune(0):
	// 	// 	return false
	// 	// case unicode.In(c, unicode.Letter):
	// 	// 	lastQuote = c
	// 	// 	return false
	// 	default:
	// 		return false
	// 		// return unicode.IsLetter(c)
	// 	}
	// }

	// // splitting string by space but considering quoted section
	// items := strings.FieldsFunc(foo, f)

	// // create and fill the map
	// m := make(map[string]string)
	// for _, item := range items {
	// 	x := strings.Split(item, "=")
	// 	m[x[0]] = x[1]
	// }

	// // print the map
	// for k, v := range m {
	// 	fmt.Printf("%s: %s\n", k, v)
	// }
}
