---
title: "Faroe.getUsers()"
---

# Faroe.getUsers()

Mapped to [GET /users](/api-reference/rest/endpoints/get_users).

Gets an array of users. Returns an empty array if there are no users.

## Definition

```ts
//$ FaroeUser=/api-reference/js/main/FaroeUser
async function getUsers(
    sortBy: UserSortBy,
    sortOrder: SortOrder,
    count: number,
    page: number
): Promise<$$FaroeUser[]>
```

### Parameters

- `sortBy`
- `sortOrder`
- `count`
- `page`

## Error codes

- `UNKNOWN_ERROR`

## Examples

```ts
import { Faroe, UserSortBy, SortOrder } from "@faroe/sdk";

const faroe = new Faroe(url, secret);

const users = await faroe.getUsers(UserSortBy.CreatedAt, SortOrder.Ascending, 20, 2);
```
