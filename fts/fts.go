package fts

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/heyvito/docuowl/fs"
	"github.com/heyvito/docuowl/fts/lang"
)

//go:generate go run ../generators/fts-stopwords.go
//go:generate go run ../generators/fts-special.go

const MinWordLength = 3

func wordLen(input string) int {
	return utf8.RuneCountInString(input)
}

func stopWords(langName string) []string {
	if l, ok := lang.StopWords[langName]; ok {
		return l
	}
	return []string{}
}

type FullTextSearchEngine struct {
	lang      string
	stopWords []string

	words          map[int]map[string]int
	pages          map[int]map[string][][]int
	frequencies    map[int]map[string]int
	sectionIndexes []string
}

func New(lang string) *FullTextSearchEngine {
	return &FullTextSearchEngine{
		lang:           lang,
		stopWords:      stopWords(lang),
		words:          map[int]map[string]int{},
		pages:          map[int]map[string][][]int{},
		sectionIndexes: []string{},
	}
}

func (fts *FullTextSearchEngine) Serialize() (string, error) {
	compressionData := []byte{0x02}
	for _, p := range fts.sectionIndexes {
		compressionData = append(compressionData, []byte(p)...)
		compressionData = append(compressionData, 0x00)
	}
	index, err := fts.serializeIndex()
	if err != nil {
		return "", err
	}
	compressionData = append(compressionData, 0x03)
	compressionData = append(compressionData, index...)

	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	_, err = w.Write(compressionData)
	if err != nil {
		return "", err
	}
	if err = w.Close(); err != nil {
		return "", err
	}

	compressedData := buf.Bytes()
	result := []byte{0x6f, 0x77, 0x6c, 0x00, 0x01}
	result = append(result, makeUint32(uint32(len(compressedData)))...)
	result = append(result, compressedData...)

	return base64.StdEncoding.EncodeToString(result), nil
}

func makeUint16(v uint16) []byte {
	return []byte{
		byte((v >> 8) & 0xFF),
		byte(v & 0xFF),
	}
}

func makeUint32(v uint32) []byte {
	return []byte{
		byte((v >> 24) & 0xFF),
		byte((v >> 16) & 0xFF),
		byte((v >> 8) & 0xFF),
		byte(v & 0xFF),
	}
}

func (fts *FullTextSearchEngine) serializeIndex() ([]byte, error) {
	var result []byte
	for k, v := range fts.pages {
		result = append(result, uint8(k))
		result = append(result, uint8(len(v)))
		for w, ps := range v {
			result = append(result, []byte(w)...)
			result = append(result, uint8(len(ps)))
			for _, pidAndFreq := range ps {
				pid := pidAndFreq[0]
				freq := pidAndFreq[1]
				result = append(result, makeUint16(uint16(pid))...)
				result = append(result, makeUint16(uint16(freq))...)
			}
		}
	}
	return result, nil
}

func (fts *FullTextSearchEngine) AddSection(sec *fs.Section) {
	identifier := slugifySectionName(sec)
	if identifier == "" {
		return
	}
	frequency := fts.processSectionWords(sec)
	fts.sectionIndexes = append(fts.sectionIndexes, identifier)
	idx := len(fts.sectionIndexes) - 1

	for k, v := range frequency {
		wordLen := wordLen(k)
		wordsByLen, ok := fts.pages[wordLen]
		if !ok {
			wordsByLen = map[string][][]int{}
		}

		pagesForWord, ok := wordsByLen[k]
		pagesForWord = append(pagesForWord, []int{idx, v})
		wordsByLen[k] = pagesForWord
		fts.pages[wordLen] = wordsByLen
	}
}

var tokenizerStripper = regexp.MustCompile(`[\r\n\t]`)
var onlyDigits = regexp.MustCompile(`^[0-9]+$`)

func (fts *FullTextSearchEngine) isStopWord(str string) bool {
	for _, w := range fts.stopWords {
		if str == w {
			return true
		}
	}
	return false
}

func (fts *FullTextSearchEngine) tokenize(str string) []string {
	str = tokenizerStripper.ReplaceAllString(str, " ")
	wordList := strings.Split(str, " ")
	wordCount := 0
	for i, w := range wordList {
		if onlyDigits.MatchString(w) || wordLen(w) <= MinWordLength || fts.isStopWord(w) {
			wordList[i] = ""
			continue
		}
		wordCount++
	}

	newWords := make([]string, 0, wordCount)
	for _, w := range wordList {
		if w != "" {
			newWords = append(newWords, w)
		}
	}

	return newWords
}

func slugifySectionName(sec *fs.Section) string {
	if sec.Meta() == nil {
		return ""
	}
	takeID := func(entity fs.Entity) string {
		if entity.Meta().ID != "" {
			return entity.Meta().ID
		} else {
			return entity.Meta().Title
		}
	}

	ids := []string{takeID(sec)}

	p := sec.Parent
	for p != nil {
		ids = append(ids, takeID(p))
		p = p.Parent
	}

	for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
		ids[i], ids[j] = ids[j], ids[i]
	}
	return strings.Join(ids, "-")
}

func (fts *FullTextSearchEngine) processSectionWords(sec *fs.Section) map[string]int {
	data := strings.Join(sec.Content, " ")
	if sec.HasSideNotes {
		data = data + strings.Join(sec.SideNotes, " ")
	}
	data = lang.SpecialsRegexp.ReplaceAllString(data, " ")
	data = strings.ToLower(data)
	rawTokens := fts.tokenize(data)
	freq := map[string]int{}

	for _, tok := range rawTokens {
		v, _ := freq[tok]
		v++
		freq[tok] = v
	}

	for k, v := range freq {
		wordLength := wordLen(k)
		wordsCount, ok := fts.words[wordLength]
		if !ok {
			wordsCount = map[string]int{}
		}

		wordFrequency, _ := wordsCount[k]
		wordFrequency += v
		wordsCount[k] = wordFrequency
		fts.words[wordLength] = wordsCount
	}

	return freq
}
