---
title: "CLI reference"
---

# CLI reference

## generate-secret

Generates a random secret with 200 bits of entropy using a cryptographically secure source. 

```
faroe generate-secret
```

## serve

Creates a `faroe_data` directory with an SQLite database file if it doesn't already exist and starts the server on port 4000.

```
faroe serve [...options]
```

### Options

- `--port`: The port number (default: 4000).
- `--dir`: The path of the directory to store data (default: `faroe_data`). 
- `--secret`: A random secret. If provided, requires requests to the server to include the secret in the `Authorization` header.

### Example

```
faroe serve --port=3000 --dir="/data/faroe" --secret=SECURE_SECRET
```