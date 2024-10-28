---
title: "Getting started"
---

# Getting started

Install the latest version of Faroe:

- [Download Faroe v0.1.0 for Linux (x64)](https://github.com/faroedev/faroe/releases/download/v0.2.0/linux-amd64.zip)
- [Download Faroe v0.1.0 for Linux (ARM64)](https://github.com/faroedev/faroe/releases/download/v0.2.0/linux-arm64.zip)
- [Download Faroe v0.1.0 for MacOS (x64)](https://github.com/faroedev/faroe/releases/download/v0.2.0/darwin-amd64.zip)
- [Download Faroe v0.1.0 for MacOS (ARM64)](https://github.com/faroedev/faroe/releases/download/v0.2.0/darwin-arm64.zip)
- [Download Faroe v0.1.0 for Windows (x64)](https://github.com/faroedev/faroe/releases/download/v0.2.0/windows-amd64.zip)
- [Download Faroe v0.1.0 for Windows (ARM64)](https://github.com/faroedev/faroe/releases/download/v0.2.0/windows-arm64.zip)

You can immediately start the server on port 4000 with `faroe server`:

```
./faroe serve

./faroe serve --port=3000
```

This will create a `faroe_data` folder in the root that contains the SQLite database. Remember to add this to `.gitignore`.

For production apps, generate a secret with the `generate-secret` command and pass it when starting the sever.

```
./faroe generate-secret
```

```
./faroe serve --secret=SECRET
```

You can get a formatted list of users by sending a GET request to `/users` with the `Accept` header set to `text/plain`.

```
curl http://localhost:4000/users -H "Accept: text/plain"
```