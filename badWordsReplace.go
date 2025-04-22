package main

import "strings"

func badWordsReplace(text string) string {
	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	textWords := strings.Split(text, " ")

	for i, wrd := range textWords {
		if _, ok := badWords[strings.ToLower(wrd)]; ok {
			textWords[i] = "****"
		}
	}

	newText := strings.Join(textWords, " ")

	return newText
}
