package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/joho/godotenv"
)

type BookExistResponse struct {
	XMLName xml.Name `xml:"response"`
	Request Request  `xml:"request"`
	Result  Result   `xml:"result"`
}

type Request struct {
	Isbn    string `xml:"isbn13"`
	LibCode string `xml:"libCode"`
}

type Result struct {
	HasBook       string `xml:"hasBook"`
	LoanAvailable string `xml:"loanAvailable"`
}

type LibraryInfo struct {
	LibCode   string
	Latitude  string
	Longitude string
}

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

	// apiURL에 dynamodb에서 받아온 libCode랑 프론트에서 받아온 isbn으로 loan 반환값이 Y인지 확인하고
	// Y인 배열만 모아서 프론트로 전달
	// 이때 이 배열 안에는 libCode, libName, latitude, longitude가 전달되어야 함
	var loanAvailableLibraries []LibraryInfo
	count := 0
	for _, item := range result.Items {
		if *item["libCode"].S == "129226" {
			break
		}
		if callAPI(item["libCode"], "9791191056556") { // isbn 프론트에서 받아오는 코드로 변경해야 함
			libInfo := LibraryInfo{
				LibCode:   *item["libCode"].S,
				Latitude:  *item["latitude"].S,
				Longitude: *item["longitude"].S,
			}
			loanAvailableLibraries = append(loanAvailableLibraries, libInfo)
		}
		count++
	}

	fmt.Println(loanAvailableLibraries)
	fmt.Println(count)
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
		return nil, err
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

func callAPI(libCode *dynamodb.AttributeValue, isbn string) bool {
	authKey := os.Getenv("AUTH_KEY")
	libCodeStr := *libCode.S
	apiURL := fmt.Sprintf("https://data4library.kr/api/bookExist?authKey=%s&libCode=%s&isbn13=%s", authKey, libCodeStr, isbn)

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
