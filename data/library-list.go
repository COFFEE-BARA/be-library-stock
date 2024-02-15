package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/joho/godotenv"
)

type LibraryResponse struct {
	XMLName   xml.Name  `xml:"response"`
	Request   Request   `xml:"request"`
	PageNo    int       `xml:"pageNo"`
	PageSize  int       `xml:"pageSize"`
	NumFound  int       `xml:"numFound"`
	ResultNum int       `xml:"resultNum"`
	Libraries Libraries `xml:"libs"`
}

type Request struct {
	PageNo   int `xml:"pageNo"`
	PageSize int `xml:"pageSize"`
}

type Libraries struct {
	Libraries []Library `xml:"lib"`
}

type Library struct {
	LibCode   int    `xml:"libCode"`
	LibName   string `xml:"libName"`
	Latitude  string `xml:"latitude"`
	Longitude string `xml:"longitude"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	})
	if err != nil {
		log.Fatal("Error creating session:", err)
	}

	svc := dynamodb.New(sess)

	authKey := os.Getenv("AUTH_KEY")
	fmt.Println(authKey)
	pageNo := 1
	pageSize := 30
	apiURL := fmt.Sprintf("https://data4library.kr/api/libSrch?authKey=%s&pageSize=%d", authKey, pageSize)

	for {
		response, err := http.Get(apiURL + "&pageNo=" + strconv.Itoa(pageNo))
		if err != nil {
			fmt.Println("HTTP 요청 오류:", err)
			return
		}
		defer response.Body.Close()

		byteValue, _ := ioutil.ReadAll(response.Body)

		var libraryResponse LibraryResponse

		err = xml.Unmarshal(byteValue, &libraryResponse)
		if err != nil {
			fmt.Println("XML 파싱 오류:", err)
			return
		}
		fmt.Println(libraryResponse)

		for _, lib := range libraryResponse.Libraries.Libraries {
			input := &dynamodb.PutItemInput{
				Item: map[string]*dynamodb.AttributeValue{
					"libCode": {
						S: aws.String(strconv.Itoa(lib.LibCode)),
					},
					"libName": {
						S: aws.String(lib.LibName),
					},
					"latitude": {
						S: aws.String(lib.Latitude),
					},
					"longitude": {
						S: aws.String(lib.Longitude),
					},
				},
				TableName: aws.String(os.Getenv("TABLE_NAME")),
			}

			_, err = svc.PutItem(input)
			if err != nil {
				fmt.Println("Error writing to DynamoDB:", err)
				return
			}
		}

		if libraryResponse.PageNo*pageSize >= libraryResponse.NumFound {
			break
		}

		pageNo++
	}

	fmt.Println("데이터가 성공적으로 DynamoDB에 저장되었습니다.")
}
