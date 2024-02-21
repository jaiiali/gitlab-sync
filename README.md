# Gitlab Sync

Sync your local workspace with private Gitlab repositories.

## Gitlab Token

https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html

## Usage

```shell
BASE_PATH="/workspace/git.example.com" BASE_URL="https://git.example.com" TOKEN="glpat-xxx" ORDER_BY="last_activity_at"  go run main.go
```

### cURL

```shell
curl -X GET "https://git.example.com/api/v4/projects?order_by=last_activity_at&sort=desc&per_page=100&page=1" -H "PRIVATE-TOKEN: glpat-xxx" | jq '.[] | {id, created_at, updated_at, last_activity_at, path_with_namespace, ssh_url_to_repo}' | jq '.id' | wc -l
```
