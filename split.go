package pairs

import (
	"encoding/json"
	"log"
	"os"
	"fmt"
	"errors"
)

type pair struct {
	sender   string `json:`
	receiver string `json:`
}

func (p pair) Sender() string {
	return p.sender
}

func (p pair) Receiver() string {
	return p.receiver
}

func loadPairs(filepath string) []pair {
	var pairs []pair
	file, e := os.Open(filepath)
	if e != nil {
		log.Println(e)
	}
	var data map[string]interface{}
	if e = json.NewDecoder(file).Decode(&data); e != nil {
		log.Println(e)
	}
	for k, v := range data {
		pairs = append(pairs, pair{
			sender:   k,
			receiver: v.(string),
		})
	}
	return pairs
}

func GetPair(path string, index int) (string, string, error){
	pairs := loadPairs(path)
	fmt.Printf("%v", pairs)
	if index >= len(pairs){
		e := errors.New("Pairs.GetPair : index bigger than the number of pairs %d")
		return "","",e
	}

	return pairs[index].sender, pairs[index].receiver, nil
}
