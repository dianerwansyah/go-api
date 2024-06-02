# Go example projects go-api with PostgreSQL

## GO VERSION
go version go1.19.4


## Clone the project

```
$ git clone https://github.com/dianerwansyah/go-api.git
$ cd go-api
```

## Install Database PostgreSQL
```
https://www.postgresql.org/download/
```

## Config you Database PostgreSQL
```
$ cd devops/local/config.yaml
Add or update the following settings according to the PostgreSQL database settings
```
Database, Table, Field Auto Created

## Run project

```
$ go run app/main.go
```
## API URL
```
Create Post > MethodPost : localhost:8081/api/posts
Update Post > MethodUpdate : localhost:8081/api/posts/{$id}
Delete Post > MethodDelete : localhost:8081/api/posts/{$id}
GetbyID Post > MethodGet : localhost:8081/api/posts/{$id}

Create Tag > MethodPost : localhost:8081/api/tag
Update Tag > MethodUpdate : localhost:8081/api/tag/{$id}
Delete Tag > MethodDelete : localhost:8081/api/tag/{$id}
GetbyID Tag > MethodGet : localhost:8081/api/tag/{$id}
```
## JSON

```
//Create Post
{
	"title": "Contoh Post",
	"content": "Isi dari contoh post",
	"tags": [
		{
			"label": "Go"
		},
		{
			"label": "API"
		}
	],
	"status": "Draft",
	"publish_dte": "2024-06-02T00:00:00Z"
}

//Create Tag
{
	"label": "Go 1.19.4",
}
```
