# Go example projects go-api with PostgreSQL

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
	]
}
```
