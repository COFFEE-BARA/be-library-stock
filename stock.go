package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/joho/godotenv"
)

type BookExistResponse struct {
	XMLName xml.Name `xml:"response"`
	Result  Result   `xml:"result"`
}

type Result struct {
	LoanAvailable string `xml:"loanAvailable"`
}

type LibraryInfo struct {
	LibCode   string
	Latitude  string
	Longitude string
}

// apiURL에 dynamodb에서 받아온 libCode랑 프론트에서 받아온 isbn으로 loan 반환값이 Y인지 확인하고
// Y인 배열만 모아서 프론트로 전달
// 이때 이 배열 안에는 libCode, libName, latitude, longitude가 전달되어야 함

func main() {
	loadEnv()

	sess, err := createNewSession()
	if err != nil {
		log.Fatal("Error creating session:", err)
	}

	result, err := scanDynamoDB(sess)
	if err != nil {
		log.Fatal(err)
	}

	var libraries []LibraryInfo
	for _, item := range result.Items {
		libInfo := LibraryInfo{
			LibCode:   *item["libCode"].S,
			Latitude:  *item["latitude"].S,
			Longitude: *item["longitude"].S,
		}
		libraries = append(libraries, libInfo)
	}

	var lib []LibraryInfo
	lib = callAPIs(libraries, "9788956609959")

	fmt.Println(len(lib))
	for _, info := range lib {
		fmt.Printf("%s | ", info.LibCode)
	}
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func createNewSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning createNewSession: %v", err)
	}
	return sess, nil
}

func scanDynamoDB(sess *session.Session) (*dynamodb.ScanOutput, error) {
	svc := dynamodb.New(sess)
	tableName := os.Getenv("TABLE_NAME")

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := svc.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("error scanning table: %v", err)
	}

	return result, nil
}

func callAPI(libCode string, isbn string) bool {
	authKey := os.Getenv("AUTH_KEY")
	apiURL := fmt.Sprintf("https://data4library.kr/api/bookExist?authKey=%s&libCode=%s&isbn13=%s", authKey, libCode, isbn)

	response, err := http.Get(apiURL)
	if err != nil {
		log.Fatal("Error fetching data from API:", err)
	}
	defer response.Body.Close()

	byteValue, _ := ioutil.ReadAll(response.Body)

	var bookExistResponse BookExistResponse

	err = xml.Unmarshal(byteValue, &bookExistResponse)

	if err != nil {
		log.Fatal("Error parsing XML:", err)
	}

	return bookExistResponse.Result.LoanAvailable == "Y"

}

func callAPIs(libraries []LibraryInfo, isbn string) []LibraryInfo {
	ch := make(chan LibraryInfo)

	var wg sync.WaitGroup
	for _, library := range libraries {
		wg.Add(1)

		go func(libCode string) {
			defer wg.Done()

			if callAPI(library.LibCode, isbn) {
				ch <- LibraryInfo{LibCode: libCode}
			}
		}(library.LibCode)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var loanAvailableLibraries []LibraryInfo
	for libInfo := range ch {
		if libInfo.LibCode != "" {
			loanAvailableLibraries = append(loanAvailableLibraries, libInfo)
		}
	}

	return loanAvailableLibraries
}
