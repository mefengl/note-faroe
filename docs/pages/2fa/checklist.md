---
title: "Implementation checklist"
---

# Implementation checklist

- Are users with second factors required to verify their second factor before resetting their passwords?
- Are non-2FA verified users blocked from privileged actions, including changing passwords, viewing the recovery code, registering a new TOTP credential, and generating a new recovery code? If 2FA is optional, do users without a registered second factor have access to privileged actions?
- Are users safe from being locked out from their account when your application's database errors during a query?