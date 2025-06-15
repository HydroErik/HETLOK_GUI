package apiCall

import (
	"encoding/json"
	"errors"
	"fmt"
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

type Client struct {
	ClientId        float64 `json:"clientId"`
	Name            string  `json:"name"`
	ShortName       string  `json:"shortName"`
	Notes           string  `json:"notes"`
	IsEnabled       bool    `json:"isEnabled"`
	CreateTimeStamp string  `json:"createTimestamp"`
	TimeZoneCode    float64 `json:"timezoneId"`
}

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


// Query DB api call return interface object that needs to be parsed by the function that calls
// 3 query types {oneRecordName | oneRecordId | allRecords}
// need name or ID for one record calls requires table name
func QueryDB(table string, rtype string, name string, id string) ([]interface{}, error) {
	var query string
	var res map[string]interface{}

	//Swithc based on type
	switch rtype {
	case "a":
		//All records call
		query = fmt.Sprintf("%squery?table=%s&operation=allRecords", API_address, table)
	case "n":
		//Single record by name
		if name == "" {
			return []interface{}{}, errors.New("to request by name we need the record name")
		}
		query = fmt.Sprintf("%squery?table=%s&operation=oneRecordByName&name=%s", API_address, table, name)
	case "i":
		//Single record by ID
		if name == "" {
			return []interface{}{}, errors.New("to request by ID we need the record id")
		}
		query = fmt.Sprintf("%squery?table=%s&operation=oneRecordById&recordId=%s", API_address, table, id)
	default:
		return []interface{}{}, errors.New("invalid request type: " + rtype + "types are: (a)ll, (n)ame, (i)d")
	}

	r, err := http.Get(query)
	if err != nil {
		return []interface{}{}, fmt.Errorf("failed get query: %s\nerror message: %s", query, err.Error())
	}

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		return []interface{}{}, fmt.Errorf("failed to decode JSON error: %s query: %s ", err.Error(), query)
	}

	content, ok := res["queryReply"].(map[string]interface{})["content"].([]interface{})
	if !ok {
		return []interface{}{}, fmt.Errorf("error parsing res for query reply and conten\nRes:\n%s", res)
	}
	return content, nil

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


	cl, err := GetClients()
	if err != nil {
		log.Fatalf("Failed Client call error:%s", err.Error())
	}
	fmt.Println(cl)

	cl2, err := QueryDB("client", "a", "", "")
	if err != nil {
		log.Fatalf("Failed query call with error:\n%s", err.Error())
	}

	for _, cli := range(cl2){
		cliMap := cli.(map[string]interface{})
		fmt.Printf("clientId: %d\nName: %s\nNotes: %s\n\n", int(cliMap["clientId"].(float64)), cliMap["name"].(string), cliMap["notes"].(string))

	}

}
*/
