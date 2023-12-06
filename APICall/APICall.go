package apiCall

import (
	"encoding/json"
	"fmt"
	"io"

	//"log"
	"net/http"
)

type PipeObj struct {
	Pipes []struct {
		AdminProps []struct {
			Name              string `json:"name"`
			SelectListName    string `json:"selectListName"`
			ValidationRules   string `json:"validationRules"`
			Ordering          string `json:"ordering"`
			DefaultValue      string `json:"defaultValue"`
			DataType          string `json:"dataType"`
			HelpInfo          string `json:"helpInfo"`
			DeactivationRules string `json:"deactivationRules"`
			Title             string `json:"title"`
			Value             string `json:"value"`
		} `json:"adminProperties"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Id          string `json:"id"`
		PipeType    string `json:"pipeType"`
	} `json:"pipes"`
}

// This is a test of the ability to push and pull with and app password
// Transformer API Call returns Transformer Pipe Obj or error if failed
// Note the struct to recieve response object is static in design
func TransformerCall() (PipeObj, error) {
	urlString := "http://35.88.227.145:8080/pipes?pipeType=TRANSFORMER&adminProperties=Yes"

	res, err := http.Get(urlString)
	if err != nil {
		return PipeObj{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return PipeObj{}, err
	}

	var p1 PipeObj

	err = json.Unmarshal(body, &p1)
	if err != nil {
		return PipeObj{}, err
	}

	for _, p := range p1.Pipes {
		fmt.Println(p.Name)
		for _, q := range p.AdminProps {
			fmt.Printf("\t%s\n", q)
		}
	}

	//fmt.Println(p1.Pipes[0].AdminProps[0].Name)
	return p1, nil

}

/* //Testing Main Method to test different function calls
func main() {
	pipe, err := TransformerCall()

	if err != nil{
		log.Fatalf("%s", err)
	}

	fmt.Print(pipe)
}
*/
