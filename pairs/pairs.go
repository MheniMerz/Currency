package pairs

import (
	"errors"
	"encoding/json"
	"os"
	"fmt"
)

type pair struct {
	sender   string
	receiver string
}

type line struct {
	sender    string
	receivers []string
}

type testFile struct {
	lines []line
}
//returns one pair sender/receiver
func (tf testFile) get(row int, col int) *pair {
	if row >= len(tf.lines) {
		return nil
		}
	if col >= len(tf.lines[row].receivers){
		return nil
	}
	return &pair{
		sender: tf.lines[row].sender,
		receiver: tf.lines[row].receivers[col],
	}
}
// the behavior that is executed when Decode is called
func (tf *testFile) UnmarshalJSON(data []byte) error {
	var d map[string]interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	for k, v := range d {
		l := line{
			sender: k,
		}

		for _, recv := range v.([]interface{}) {
			l.receivers = append(l.receivers, recv.(string))
		}

		tf.lines = append(tf.lines, l)
	}
	return nil
}

//reads from the json file and returns lines of possible pairs, one sender can have multiple receivers
func parseFile(file string) (*testFile, error) {
	d, e := os.Open(file)
	if e != nil {
		return nil, e
	}
	var tf testFile
	err := json.NewDecoder(d).Decode(&tf)
	if err != nil {
		return nil, err
	}
	return &tf, nil
}

func GetPair(path string, row int, col int) (string, string, error){

	tf, err := parseFile(path)
	if err != nil {
		return "","",err
	}
	fmt.Printf("%d [%d-%d] -- Pairs : ----- %v ----\n",len(tf.lines), row, col, tf)
	if row >= len(tf.lines) {
		e := errors.New("Pairs.GetPair : number of rows too big %d")
		return "","",e
	}
	if col >= len(tf.lines[row].receivers){
		e := errors.New("Pairs.GetPair : number of columns too big")
		return "","",e
	}
	pair := tf.get(row, col)
	fmt.Printf("Returned Pair : ---- %v ----", pair)
	return pair.sender, pair.receiver, nil
}
