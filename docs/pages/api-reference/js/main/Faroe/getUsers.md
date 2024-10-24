---
title: "Faroe.getUsers()"
---

# Faroe.getUsers()

Mapped to [GET /users](/api-reference/rest/endpoints/get_users).

Gets an array of users. Returns an empty array if there are no users.

## Definition

```ts
//$ PaginationResult=/api-reference/js/main/PaginationResult
//$ FaroeUser=/api-reference/js/main/FaroeUser
async function getUsers(options?: {
    sortBy: UserSortBy = UserSortBy.CreatedAt,
    sortOrder: SortOrder = SortOrder.Ascending,
    perPage: number = 20,
    page: number = 1,
    emailQuery?: string
}): Promise<$$PaginationResult<$$FaroeUser>>
```

### Parameters

- `options.sortBy` 
- `options.sortOrder`
- `options.perPage`
- `options.page`
- `options.emailQuery: A non-empty string. If defined, only users with an email that includes the keyword will be returned. Multiple keywords are not supported.

## Error codes

- `UNKNOWN_ERROR`

## Examples

```ts
import { Faroe, UserSortBy, SortOrder } from "@faroe/sdk";

const faroe = new Faroe(url, secret);

const users = await faroe.getUsers({
    sortBy: UserSortBy.CreatedAt,
    sortOrder: SortOrder.Ascending,
    perPage: 20,
    page: 2,
    emailQuery: "@example.com"
});
```
