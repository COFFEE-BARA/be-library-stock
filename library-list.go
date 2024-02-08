package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type LibraryResponse struct {
	XMLName   xml.Name `xml:"response"`
	Request   Request
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
	authKey := "cd8788b75f612015c9aa389baafaebd129dbcea9c7c4b30f3ec0b1c98bbac570"
	pageNo := 1
	pageSize := 30
	apiURL := fmt.Sprintf("https://data4library.kr/api/libSrch?authKey=%s&pageSize=%d", authKey, pageSize)

	file, err := os.Create("library-stock.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Library Code", "Library Name", "Latitude", "Longitude"}
	writer.Write(headers)

	for {
		response, err := http.Get(apiURL + "&pageNo=" + strconv.Itoa(pageNo))
		if err != nil {
			fmt.Println("HTTP 요청 오류:", err)
			return
		}
		defer response.Body.Close()

		byteValue, _ := ioutil.ReadAll(response.Body)

		var libraryResponse LibraryResponse

		// XML 언마샬링 (XML을 구조체로 파싱)
		err = xml.Unmarshal(byteValue, &libraryResponse)
		if err != nil {
			fmt.Println("XML 파싱 오류:", err)
			return
		}

		for _, lib := range libraryResponse.Libraries.Libraries {
			row := []string{
				strconv.Itoa(lib.LibCode),
				lib.LibName,
				lib.Latitude,
				lib.Longitude,
			}
			writer.Write(row)
		}

		if libraryResponse.PageNo*pageSize >= libraryResponse.NumFound {
			break
		}

		pageNo++
	}

	fmt.Println("CSV 파일이 성공적으로 생성되었습니다.")
}
