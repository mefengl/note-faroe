---
title: "Faroe.getUsers()"
---

# Faroe.getUsers()

Mapped to [GET /users](/reference/rest/endpoints/get_users).

Gets an array of users. Returns an empty array if there are no users.

## Definition

```ts
//$ PaginationResult=/reference/sdk-js/main/PaginationResult
//$ FaroeUser=/reference/sdk-js/main/FaroeUser
async function getUsers(options?: {
    sortBy: UserSortBy = UserSortBy.CreatedAt,
    sortOrder: SortOrder = SortOrder.Ascending,
    perPage: number = 20,
    page: number = 1
}): Promise<$$PaginationResult<$$FaroeUser>>
```

### Parameters

- `options.sortBy` 
- `options.sortOrder`
- `options.perPage`
- `options.page`

## Error codes

- `INTERNAL_ERROR`

## Examples

```ts
import { Faroe, UserSortBy, SortOrder } from "@faroe/sdk";

const faroe = new Faroe(url, secret);

const users = await faroe.getUsers({
    sortBy: UserSortBy.CreatedAt,
    sortOrder: SortOrder.Ascending,
    perPage: 20,
    page: 2
});
```
