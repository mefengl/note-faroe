---
title: "FaroeError"
---

# FaroeError

Extends `Error`.

An error indicating a server error response.

## Properties

```ts
interface Properties {
    status: number;
    code: string;
}
```

- `status`: HTTP response status.
- `code`: Faroe error code.
