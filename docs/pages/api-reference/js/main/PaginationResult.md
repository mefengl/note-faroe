---
title: "PaginationResult"
---

# PaginationResult

Represents a paginated result.

## Definition

```ts
interface PaginationResult<T> {
    totalPages: number;
    items: T[]
}
```

### Type parameters

- `T`

### Properties

- `totalPages`
- `items`: Items in the page.