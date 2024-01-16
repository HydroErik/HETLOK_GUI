package apicall

import (
	"encoding/json"
	"io"

	//"log"
	"net/http"
)

const API_address = "http://localhost:8080/"

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

type ListsObj struct {
	Lists []struct {
		ListId           string `json:"selectListId"`
		ListName         string `json:"selectListName"`
		ListDescription  string `json:"selectListDescription"`
		ListDefaultName  string `json:"selectListDefaultName"`
		ListDefaultValue string `json:"selectListDefaultValue"`
		ListItems        []struct {
			Ordering string `json:"ordering"`
			Name     string `json:"name"`
			Value    string `json:"value"`
		} `json:"selectListItems"`
	} `json:"selectLists"`
}

// This is a test of the ability to push and pull with and app password
// Transformer API Call returns Transformer Pipe Obj or error if failed
// Note the struct to recieve response object is static in design
func TransformerCall(admin bool) (PipeObj, error) {
	var urlString string

	if admin {
		urlString = API_address + "pipes?pipeType=TRANSFORMER&adminProperties=Yes"
	} else {
		urlString = API_address + "pipes?pipeType=TRANSFORMER&adminProperties=No"
	}
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

	/*
		for _, p := range p1.Pipes {
			fmt.Println(p.Name)
			for _, q := range p.AdminProps {
				//fmt.Printf("\t%s\n", q)
			}
		}
	*/

	return p1, nil

}

// select list API call
// id = string list ID number default = 0
// noItem = string weather to add list items in response default = true
// returns ListsObj struct
func GetLists(id string, noItem string) (ListsObj, error) {
	urlString := API_address + "selectLists?selectList=" + id + "&noItems=" + noItem

	res, err := http.Get(urlString)
	if err != nil {
		return ListsObj{}, nil
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ListsObj{}, err
	}

	var l ListsObj

	err = json.Unmarshal(body, &l)
	if err != nil {
		return ListsObj{}, err
	}

	return l, nil

}

/*
// Testing Main Method to test different function calls
func main() {
	/*
		pipe, err := TransformerCall(true)

		if err != nil {
			log.Fatalf("%s", err)
		}
		fmt.Printf("\n\n\n\n\n\n%s", pipe)


	/*
		lists, err := GetLists("0", "false")

		if err != nil {
			log.Fatalf("%s", err)
		}

		fmt.Printf("\n\n\n\n\n\n%s", lists)


	fmt.Println(string(0))


}
*/
