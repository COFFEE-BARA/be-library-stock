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
	"github.com/aws/aws-lambda-go/lambda"
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

type Location struct {
	Latitude  float64
	Longitude float64
}

// apiURL에 dynamodb에서 받아온 libCode랑 프론트에서 받아온 isbn으로 loan 반환값이 Y인지 확인하고
// Y인 배열만 모아서 프론트로 전달
// 이때 이 배열 안에는 libCode, libName, latitude, longitude가 전달되어야 함

func EventHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	latStr := request.QueryStringParameters["lat"]
	lonStr := request.QueryStringParameters["lon"]

	lat, _ := strconv.ParseFloat(latStr, 64)
	lon, _ := strconv.ParseFloat(lonStr, 64)

	location := Location{
		Latitude:  lat,
		Longitude: lon,
	}

	loadEnv()

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

	libraries := libraryHandler(result, location)

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

func libraryHandler(result *dynamodb.ScanOutput, location Location) []LibraryInfo {
	var libraries []LibraryInfo
	for _, item := range result.Items {
		distance := calculateDistance(location, *item["latitude"].S, *item["longitude"].S)

		if distance <= 30 {
			libInfo := LibraryInfo{
				LibCode:   *item["libCode"].S,
				Latitude:  *item["latitude"].S,
				Longitude: *item["longitude"].S,
			}
			libraries = append(libraries, libInfo)
		}

	}
	fmt.Println("len(libraries): ", len(libraries))

	var lib []LibraryInfo
	lib = callAPIs(libraries, "9788956609959")

	// 프론트엔드에 전달할 라이브러리 정보 반환
	return lib
}

func calculateDistance(location Location, latitude string, longitude string) float64 {
	lat1 := location.Latitude
	lon1 := location.Longitude
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

func main() {
	lambda.Start(EventHandler)
	// http.HandleFunc("/api/book/9788956609959/lending-library", receiveHandler)

	// log.Fatal(http.ListenAndServe(":8000", nil))
}

// func receiveHandler(w http.ResponseWriter, r *http.Request) {
// 	latStr := r.FormValue("lat")
// 	lonStr := r.FormValue("lon")

// 	lat, _ := strconv.ParseFloat(latStr, 64)
// 	lon, _ := strconv.ParseFloat(lonStr, 64)

// 	location := Location{
// 		Latitude:  lat,
// 		Longitude: lon,
// 	}

// 	connectDynamoDB(location)

// 	fmt.Fprintf(w, "Received location data: Latitude=%f, Longitude=%f", location.Latitude, location.Longitude)
// }

// func connectDynamoDB(location Location) {
// 	loadEnv()

// 	sess, err := createNewSession()
// 	if err != nil {
// 		log.Fatal("Error creating session:", err)
// 	}

// 	result, err := scanDynamoDB(sess)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	libraryHandler(result, location)
// }

// func libraryHandler(result *dynamodb.ScanOutput, location Location) {
// 	var libraries []LibraryInfo
// 	for _, item := range result.Items {
// 		distance := calculateDistance(location, *item["latitude"].S, *item["longitude"].S)

// 		if distance <= 30 {
// 			libInfo := LibraryInfo{
// 				LibCode:   *item["libCode"].S,
// 				Latitude:  *item["latitude"].S,
// 				Longitude: *item["longitude"].S,
// 			}
// 			libraries = append(libraries, libInfo)
// 		}

// 	}
// 	fmt.Println("len(libraries): ", len(libraries))

// 	var lib []LibraryInfo
// 	lib = callAPIs(libraries, "9788956609959")

// 	http.HandleFunc("/api/book/9788956609959/lending-library", func(w http.ResponseWriter, r *http.Request) {
// 		sendHandler(w, r, lib)
// 	})

// 	log.Fatal(http.ListenAndServe(":8000", nil))

// 	// 이제 lib 배열을 프론트엔드로 넘겨줘야함
// 	// for _, info := range lib {
// 	// 	fmt.Printf("%s | ", info.LibCode)
// 	// }
// }

// func sendHandler(w http.ResponseWriter, r *http.Request, lib []LibraryInfo) {
// 	jsonData, err := json.Marshal(lib)
// 	if err != nil {
// 		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(jsonData)
// }

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
