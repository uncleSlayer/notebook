module enterprise_search

go 1.22

require (
	github.com/joho/godotenv v1.5.1
	github.com/my-company/company-go-sdk v0.0.0
)

require golang.org/x/sync v0.8.0 // indirect

replace github.com/my-company/company-go-sdk => /home/siddhant/dev/pipeshub-go/sdk
