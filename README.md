# 📚 Checkbara
![checkbara1](https://github.com/COFFEE-BARA/be-library-stock/assets/72396865/5a3da8b6-009d-40f5-a3f8-a1561df83631)
![checkbara2](https://github.com/COFFEE-BARA/be-library-stock/assets/72396865/eeef4f87-fb54-471b-bbba-2f9fac767823)
![checkbara3](https://github.com/COFFEE-BARA/be-library-stock/assets/72396865/3a8da5c0-19ee-4251-9161-96a835ce0b3f)
<img width="875" alt="checkbara4" src="https://github.com/COFFEE-BARA/be-library-stock/assets/72396865/fe97f585-7be6-4810-bc0a-f3c331ce9c39">

<br/>

| <img width="165" alt="suwha" src="https://github.com/COFFEE-BARA/be-bookstore-stock/assets/72396865/19e01fac-5384-4ec7-98f1-9e1e613429b4"> | <img width="165" alt="yoonju" src="https://github.com/COFFEE-BARA/be-bookstore-stock/assets/72396865/fb0a14c6-2d02-4105-962e-4565663817cc"> | <img width="165" alt="yugyeong" src="https://github.com/COFFEE-BARA/be-bookstore-stock/assets/72396865/90b7268d-92e5-43d1-9da8-ae48afd9e8c1"> | <img width="165" alt="dayeon" src="https://github.com/COFFEE-BARA/be-bookstore-stock/assets/72396865/f19e65e6-0856-4b6a-a355-993ce83ddcb7"> |
| --- | --- | --- | --- |
| 🐼[유수화](https://github.com/YuSuhwa-ve)🐼 | 🐱[송윤주](https://github.com/raminicano)🐱 | 🐶[현유경](https://github.com/yugyeongh)🐶 | 🐤[양다연](https://github.com/dayeon1201)🐤 |
| Server / Data / BE | AI / Data / BE | Infra / BE / FE | BE / FE |

<br/>

# be-library-stock

<br/>

# 💿 Dynamo DB table 구조도

| libCode (PK) | libName (SK) | latitude | longitude |
| --- | --- | --- | --- |

<br/>

# 🤖 API 명세

- URL: BASE_URL/api/book/{책의 isbn 값}/library?lat={현재 위도}&lon={현재 경도}
- Method: `GET`
- 기능 소개: 해당 책을 어느 도서관에서 대출이 가능한지 지도에 표시해주는 기능

<br/>

# 🗣️ Request

## ☝🏻Request Header

```
Content-Type: application/json
```

## ✌🏻Request Params

| Name | Type | Description | Required |
| --- | --- | --- | --- |
| 책의 { isbn 값 } | String | 책의 13자리 isbn 값 | Required |

## ✌🏻Request Query

| Name | Type | Description | Required |
| --- | --- | --- | --- |
| lat={ 현재 위도 } | String | 사용자의 현재 위치 위도 | Required |
| lon={ 현재 경도 } | String | 사용자의 현재 위치 경도 | Required |

<br/>

# 🗣️ Response

## ☝🏻Response Body

```json
{
    "code": 200,
    "message": "책의 대출 가능 도서관 리스트를 가져오는데 성공했습니다.",
    "data": {
        "isbn": "9791140708116",
        "title": "아는 만큼 보이는 백엔드 개발 (한 권으로 보는 백엔드 로드맵과 커리어 가이드)",
        "libraryList": [
            {
                "code": "111110",
                "name": "구립증산정보도서관",
                "latitude": "37.582973",
                "longtitude": "126.907543"
            },
            {
                "code": "111511",
                "name": "신당누리도서관",
                "latitude": "37.5633821",
                "longtitude": "127.012303"
            }
        ]
    }
}
```

```json
{
  "code": 200,
  "message": "책의 대출 가능 도서관 리스트를 가져오는데 성공했습니다.",
    "data": {
      "isbn" : 9791140708116,
      "title" : "아는 만큼 보이는 백엔드 개발 (한 권으로 보는 백엔드 로드맵과 커리어 가이드)",
      "bookStoreList" : null
    }
}
```

## ✌🏻실패

1. 필요한 값이 없는 경우
    
    ```json
    {
      "code": 400,
      "message": "isbn값이 없습니다.",
      "data": null
    }
    ```
    
2. isbn 값에 매칭되는 책이 없을 경우
    
    ```json
    {
      "code": 404,
      "message": "없는 책입니다.",
      "data": null
    }
    ```
    
3. 서버에러
    
    ```json
    {
      "code": 500,
      "message": "서버 에러",
      "data": null
    }
    ```
    
<br/>

# 🏆 Tech Stack

## Programming Language

<img src="https://img.shields.io/badge/go-00ADD8?style=for-the-badge&logo=go&logoColor=white"/>

## DB

<img src="https://img.shields.io/badge/amazondynamodb-4053D6?style=for-the-badge&logo=amazondynamodb&logoColor=white"/>

## CI/CD & Deploy

<img src="https://img.shields.io/badge/codebuild-68A51C?style=for-the-badge&logo=codebuild&logoColor=white"/> <img src="https://img.shields.io/badge/codepipeline-527FFF?style=for-the-badge&logo=codepipeline&logoColor=white"/> <img src="https://img.shields.io/badge/docker-2496ED?style=for-the-badge&logo=docker&logoColor=white"> <img src="https://img.shields.io/badge/awslambda-FF9900?style=for-the-badge&logo=awslambda&logoColor=white"/> <img src="https://img.shields.io/badge/amazonapigateway-FF4F8B?style=for-the-badge&logo=amazonapigateway&logoColor=white"/> <img src="https://img.shields.io/badge/ecr-FC4C02?style=for-the-badge&logo=ecr&logoColor=white"/>

## Develop Tool

<img src="https://img.shields.io/badge/postman-FF6C37?style=for-the-badge&logo=postman&logoColor=white"> <img src="https://img.shields.io/badge/github-181717?style=for-the-badge&logo=github&logoColor=white"> <img src="https://img.shields.io/badge/git-F05032?style=for-the-badge&logo=git&logoColor=white">

## Communication Tool

<img src="https://img.shields.io/badge/slack-4A154B?style=for-the-badge&logo=slack&logoColor=white"> <img src="https://img.shields.io/badge/notion-000000?style=for-the-badge&logo=notion&logoColor=white">

<br/>

# 🏡 be-library-stock architecture
<img width="566" alt="be-library-stock-archi" src="https://github.com/COFFEE-BARA/be-library-stock/assets/72396865/c93b72ca-8870-4912-b5ae-30afdaa50fad">

<br/>

1. 도서관 정보나루의 도서관 정보 조회 API와 `go aws sdk`를 이용해 도서관 코드, 지점명, 위경도 값을 DynamoDB에 저장<br/>
2. API Gateway에서 요청이 들어오면 Lambda 실행 <br/>
3. Lambda는 API Gateway에서 받아온 책의 isbn 값과 DynamoDB에 저장된 데이터를 바탕으로 도서관별 도서의 대출 가능 여부 조회<br/>
4. 대출 가능한 도서관의 정보(도서관 코드, 도서관 이름, 재고 수, 위도, 경도)를 전달
