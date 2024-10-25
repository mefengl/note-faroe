---
title: "GET /users"
---

# GET /users

Gets a list of users.

```
GET https://your-domain.com/users
```

## Query parameters

All parameters are optional.

- `sort_by`: Field to sort the list by. One of:
    - `created_at` (default): Sort by when the user was created.
    - `email`:  Sort by the user's email.
    - `id`: Sort by the user's ID.
- `sort_order` Order of the list. One of:
    - `ascending` (default)
    - `descending`
- `per_page`: A positive integer that specifies the number of items in a page (default: 20).
- `page`: A positive integer that specifies the page number to be returned (default: 1).
- `email_query: A non-empty string. Only users with an email that includes the keyword will be returned. Multiple keywords are not supported.

### Example

```
/users?sort_by=created_at&sort_order=descending&per_page=50&page=2&email_query=%40example.com
```

## Successful response

Returns a JSON array of [user models](/api-reference/rest/models/user). If there are no users in the page, it will return an empty array.

You can get the number of total pages from the `X-Pagination-Total-Pages` header and the total number of users with the `X-Pagination-Total` header..

```
X-Pagination-Total-Pages: 6
X-Pagination-Total: 113
```

### Example

```json
[
    {
        "id": "eeidmqmvdtjhaddujv8twjug",
        "created_at": 1728783738,
        "email": "user@example.com",
        "email_verified": true,
        "registered_totp": false
    }
]
```

## Error codes

- [500] `UNKNOWN_ERROR`
