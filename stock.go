package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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

type Location struct {
	Latitude  string
	Longitude string
}

func main() {
	lambda.Start(EventHandler)
}

func EventHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Set CORS headers
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*", // Allow requests from any origin
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "*", // Allow all methods
		// Add more CORS headers if needed
	}

	// 1. url path paramether로 isbn 값 받아오기
	isbn, ok := request.PathParameters["isbn"]
	if !ok {
		bodyJSON, err := json.Marshal(Response{
			Code:    400,
			Message: "isbn값이 없습니다.",
		})
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Headers: headers}, err
		}

		// Return the response
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       string(bodyJSON),
		}, nil
	}

	//2. url  쿼리 파라미터 값 받아오기
	lat, latOk := request.QueryStringParameters["lat"]
	lon, lonOk := request.QueryStringParameters["lon"]
	if !(latOk && lonOk) {
		bodyJSON, err := json.Marshal(Response{
			Code:    400,
			Message: "lat 또는 lon값이 없습니다.",
		})
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Headers: headers}, err
		}

		// Return the response
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       string(bodyJSON),
		}, nil

	}

	location := Location{
		Latitude:  lat,
		Longitude: lon,
	}

	// loadEnv()

	sess, err := createNewSession()
	if err != nil {
		log.Println("Error creating session:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	result, err := scanDynamoDB(sess)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	libraries := libraryHandler(result, location, isbn)

	responseBody, err := json.Marshal(libraries)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

// func loadEnv() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// }

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

func libraryHandler(result *dynamodb.ScanOutput, location Location, isbn string) []LibraryInfo {
	var libraries []LibraryInfo
	for _, item := range result.Items {
		libCode := *item["libCode"].S
		latitude := *item["latitude"].S
		longitude := *item["longitude"].S
		fmt.Println(libCode, latitude, longitude)
		distance := calculateDistance(location, latitude, longitude)

		if distance <= 30 {
			libInfo := LibraryInfo{
				LibCode:   libCode,
				Latitude:  latitude,
				Longitude: longitude,
			}
			libraries = append(libraries, libInfo)
		}

	}

	var lib []LibraryInfo
	lib = callAPIs(libraries, isbn)

	// 프론트엔드에 전달할 라이브러리 정보 반환
	fmt.Println("---result---")
	fmt.Println(lib)
	fmt.Println(len(lib))
	return lib
}

func calculateDistance(location Location, latitude string, longitude string) float64 {
	lat1, _ := strconv.ParseFloat(location.Latitude, 64)
	lon1, _ := strconv.ParseFloat(location.Longitude, 64)
	lat2, _ := strconv.ParseFloat(latitude, 64)
	lon2, _ := strconv.ParseFloat(longitude, 64)

	const earthRadius = 6371

	lat1 = lat1 * math.Pi / 180
	lon1 = lon1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	lon2 = lon2 * math.Pi / 180

	dlon := lon2 - lon1
	dlat := lat2 - lat1
	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
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

		go func(lib LibraryInfo) {
			defer wg.Done()

			if callAPI(lib.LibCode, isbn) {
				ch <- lib
			}
		}(library)
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
