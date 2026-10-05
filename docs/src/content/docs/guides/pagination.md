---
title: Pagination
description: Use List, ListAll, and ListPages with iter.Seq2.
---

Most list endpoints expose three helpers:

| Method | Returns |
| --- | --- |
| `List` | One page |
| `ListAll` | Every item across pages (`iter.Seq2`) |
| `ListPages` | Every page envelope (`iter.Seq2`) |

```go
for charge, err := range client.Charges.ListAll(ctx, &woovi.ChargeListParams{
	Limit: 100,
}) {
	if err != nil {
		return err
	}
	_ = charge
}

for page, err := range client.Charges.ListPages(ctx, &woovi.ChargeListParams{
	Limit: 100,
}) {
	if err != nil {
		return err
	}
	log.Println(len(page.Charges), page.PageInfo.HasNextPage)
}
```
