package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
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
	"github.com/elastic/go-elasticsearch/v8"
)

type Response struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Data    *ResponseData `json:"data"`
}

type ResponseData struct {
	Isbn string `json:"isbn"`
	//Title       string        `json:"title"`
	LibraryList []LibraryInfo `json:"libraryList"`
}
type BookExistResponse struct {
	XMLName xml.Name `xml:"response"`
	Result  Result   `xml:"result"`
}

type Result struct {
	LoanAvailable string `xml:"loanAvailable"`
}

type LibraryInfo struct {
	LibCode   string `json:"code"`
	LibName   string `json:"name"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longtitude"`
}

type Location struct {
	Latitude  string
	Longitude string
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// //0. 환경변수
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// CLOUD_ID := os.Getenv("CLOUD_ID")
	// API_KEY := os.Getenv("API_KEY")
	// INDEX_NAME := os.Getenv("INDEX_NAME")
	// FIELD_NAME := os.Getenv("FIELD_NAME")

	REGION := os.Getenv("REGION")
	TABLE_NAME := os.Getenv("TABLE_NAME")

	AUTH_KEY_SH := os.Getenv("AUTH_KEY_SH")
	AUTH_KEY_YG := os.Getenv("AUTH_KEY_YG")
	AUTH_KEY_YJ := os.Getenv("AUTH_KEY_YJ")
	AUTH_KEY_DY := os.Getenv("AUTH_KEY_DY")

	// 호출이 안되면 다른 auth_key로 두기
	var authKeyList []string
	authKeyList = append(authKeyList, AUTH_KEY_SH)
	authKeyList = append(authKeyList, AUTH_KEY_YG)
	authKeyList = append(authKeyList, AUTH_KEY_YJ)
	authKeyList = append(authKeyList, AUTH_KEY_DY)

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

	//3. escloud에서 책이름 가져오기

	// esClient, err := connectElasticSearch(CLOUD_ID, API_KEY)
	// if err != nil {
	// 	fmt.Println("Error connecting to Elasticsearch:", err)
	// 	return events.APIGatewayProxyResponse{StatusCode: 500, Headers: headers}, err
	// }
	// //3.1 isbn 값으로 검색하기
	// title, err := searchTitle(esClient, INDEX_NAME, FIELD_NAME, isbn)
	// if err != nil {
	// 	fmt.Println("인덱스 검색 중 오류 발생:", err)
	// 	return events.APIGatewayProxyResponse{StatusCode: 500, Headers: headers}, err
	// }

	//4. dynamoDB에서 도서관정봅가져오기
	sess, err := createNewSession(REGION)
	if err != nil {
		log.Println("Error creating session:", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Headers: headers}, err
	}
	result, err := scanDynamoDB(sess, TABLE_NAME)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Headers: headers}, err
	}

	//5. 도서관 api 돌려서 대출가능한 도서관 가져오기
	libraries, err := libraryHandler(result, location, isbn, authKeyList)
	if err != nil {
		log.Println(err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Headers: headers}, err
	}

	bodyJson, err := json.Marshal(Response{
		Code:    200,
		Message: "책의 대출 가능 도서관 리스트를 가져오는데 성공했습니다.",
		Data: &ResponseData{
			Isbn: isbn,
			//Title:       title,
			LibraryList: libraries,
		},
	})

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(bodyJson),
	}, nil
}

func connectElasticSearch(CLOUD_ID, API_KEY string) (*elasticsearch.Client, error) {
	config := elasticsearch.Config{
		CloudID: CLOUD_ID,
		APIKey:  API_KEY,
	}

	es, err := elasticsearch.NewClient(config)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	fmt.Print("엘라스틱 클라이언트 : ", es)

	// Elasticsearch 서버에 핑을 보내 연결을 테스트합니다.
	res, err := es.Ping()
	if err != nil {
		fmt.Println("Elasticsearch와 연결 중 오류 발생:", err)
		return nil, err
	}
	defer res.Body.Close()

	fmt.Println("Elasticsearch 클라이언트가 성공적으로 연결되었습니다.")

	return es, nil

}

func searchTitle(es *elasticsearch.Client, indexName, fieldName, value string) (string, error) {

	//검색 쿼리 작성
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				fieldName: value,
			},
		},
	}

	// 쿼리를 JSON으로 변환합니다.
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return "", err
	}

	// 검색 요청을 수행합니다.
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(indexName),
		es.Search.WithBody(bytes.NewReader(queryJSON)),
	)
	if err != nil {
		return "", err
	}

	// 검색 응답을 디코딩합니다.
	var searchResponse map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		fmt.Println("검색 응답 디코딩 중 오류 발생:", err)
		return "", err
	}

	// 히트를 추출하고 후 저장
	hits := searchResponse["hits"].(map[string]interface{})["hits"].([]interface{})
	temp := hits[0].(map[string]interface{})["_source"].(map[string]interface{})

	return temp["Title"].(string), nil

}

func createNewSession(REGION string) (*session.Session, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning createNewSession: %v", err)
	}
	return sess, nil
}

func scanDynamoDB(sess *session.Session, TABLE_NAME string) (*dynamodb.ScanOutput, error) {
	svc := dynamodb.New(sess)
	tableName := TABLE_NAME

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := svc.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("error scanning table: %v", err)
	}

	return result, nil
}

func libraryHandler(result *dynamodb.ScanOutput, location Location, isbn string, authKeyList []string) ([]LibraryInfo, error) {
	var libraries []LibraryInfo
	for _, item := range result.Items {
		libCode := *item["libCode"].S
		latitude := *item["latitude"].S
		longitude := *item["longitude"].S
		libName := *item["libName"].S

		distance := calculateDistance(location, latitude, longitude)

		if distance <= 10 {
			libInfo := LibraryInfo{
				LibCode:   libCode,
				LibName:   libName,
				Latitude:  latitude,
				Longitude: longitude,
			}
			libraries = append(libraries, libInfo)

		}

	}

	//var lib []LibraryInfo
	lib, err := callAPIs(libraries, isbn, authKeyList)
	if err != nil {
		return nil, err
	}

	// 프론트엔드에 전달할 라이브러리 정보 반환
	fmt.Println("---result---")
	fmt.Println(lib)
	fmt.Println(len(lib))
	return lib, nil
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

func callAPI(libCode string, isbn string, authKeyList []string) (bool, error) {

	for _, authKey := range authKeyList {
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

		if bookExistResponse.Result.LoanAvailable == "Y" {
			return true, nil
		} else if bookExistResponse.Result.LoanAvailable == "N" {
			return false, nil
		} else {
			continue
		}

	}

	return false, errors.New("도서나루 API와 통신이 불가합니다.")

}

func callAPIs(libraries []LibraryInfo, isbn string, authKeyList []string) ([]LibraryInfo, error) {
	ch := make(chan LibraryInfo)

	var wg sync.WaitGroup
	for _, library := range libraries {
		wg.Add(1)

		go func(lib LibraryInfo) error {
			defer wg.Done()

			flag, err := callAPI(lib.LibCode, isbn, authKeyList)
			if err != nil {
				return err
			}
			if flag {
				ch <- lib
			}
			return nil
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

	return loanAvailableLibraries, nil
}

func main() {
	// 람다
	lambda.Start(handler)

	// //test~~~~~~~~~~~~~~~~~~~~~~~~~~
	// testEventFile, err := os.Open("test-event.json")
	// if err != nil {
	// 	log.Fatalf("Error opening test event file: %s", err)
	// }
	// defer testEventFile.Close()

	// // Decode the test event JSON
	// var testEvent events.APIGatewayProxyRequest
	// err = json.NewDecoder(testEventFile).Decode(&testEvent)
	// if err != nil {
	// 	log.Fatalf("Error decoding test event JSON: %s", err)
	// }

	// // Invoke the Lambda handler function with the test event
	// response, err := handler(context.Background(), testEvent)
	// if err != nil {
	// 	log.Fatalf("Error invoking Lambda handler: %s", err)
	// }

	// // Print the response
	// fmt.Printf("%v\n", response.StatusCode)
	// fmt.Printf("%v\n", response.Body)
}
