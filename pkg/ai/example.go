package ai

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"math/rand"
)

//go:embed train_data.txt
var trainDataBytes []byte
var commentExamples map[uint][]string

type TrainData struct {
	PersonaBase string `json:"persona_base"`
	Comment     string `json:"comment"`
}

func init() {
	commentExamples = make(map[uint][]string)

	personaMap := map[string]uint{
		"A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6,
	}

	scanner := bufio.NewScanner(bytes.NewReader(trainDataBytes))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var data TrainData
		if err := json.Unmarshal(line, &data); err != nil {
			continue
		}

		if aiType, exists := personaMap[data.PersonaBase]; exists {
			commentExamples[aiType] = append(commentExamples[aiType], data.Comment)
		}
	}
}

func GetRandomExamples(aiType uint, count int) []string {
	examples := commentExamples[aiType]
	total := len(examples)

	if total == 0 {
		return nil
	}

	if total <= count {
		return examples
	}

	shuffled := make([]string, total)
	copy(shuffled, examples)
	rand.Shuffle(total, func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:count]
}
