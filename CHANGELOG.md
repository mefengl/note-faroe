# Changelog

## 0.2.1

- Added `--dir` option `serve` command.˝

## 0.2.0

- Removed `email` field from user model.
    - Make sure to add a unique constraint to your user table's email field.
- Replaced endpoints:
    - `POST /authenticate/password` with `POST /users/[user_id]/verify-password`
    - `POST /update-email` with `POST /verify-new-email`
    - `POST /password-reset-requests` with `POST /users/[user_id]/password-reset-requests`
- Updated endpoint request parameters and response:
    - `GET /users`
    - `POST /users`