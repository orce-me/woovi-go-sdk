---
title: Errors
description: Handle typed APIError values from the Woovi Go SDK.
---

Failed API calls return `*woovi.APIError` when the server replies with an error body.

```go
result, err := client.Charges.Create(ctx, params)
if err != nil {
	var apiErr *woovi.APIError
	if errors.As(err, &apiErr) {
		log.Println(apiErr.StatusCode, apiErr.Message)
		return
	}
	log.Fatal(err)
}
```

Use `errors.Is` / `errors.As` instead of string matching.
