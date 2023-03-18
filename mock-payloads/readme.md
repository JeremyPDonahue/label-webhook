# Example Curl Requests

## Pod Admission

Request
```bash
curl \
  --insecure \
  --silent \
  --request POST \
  --header "Content-Type: application/json" \
  --data @./mock-payloads/pods/test-pod01.json \
  https://localhost:8443/api/v1/admit/pod
```
Response
```json
{
  "response": {
    "uid": "60df4b0b-8856-4ce7-9fb3-bc8034856995",
    "allowed": false,
    "status": {
      "metadata": {},
      "message": "You cannot use the tag 'latest' in a container."
    }
  }
}
```