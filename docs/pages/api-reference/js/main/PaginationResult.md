---
title: "PaginationResult"
---

# PaginationResult

Represents a paginated result.

## Definition

```ts
interface PaginationResult<T> {
    total: number;
    totalPages: number;
    items: T[]
}
```

### Type parameters

- `T`

### Properties

- `total`: Total number of records.
- `totalPages`
- `items`: Items in the page.