---
title: "Implementation checklist"
---

# Implementation checklist

- Do you normalize email addresses?
- Do you pass the user's IP address to all Faroe methods that accept it?
- Do you check that the password reset session has been email verified?
- Can users without a verified email address manually request for a new verification code?
- Are users without a verified email address blocked from actions that require a verified email address?
- Do you invalidate all sessions belonging to a user after they update or reset their password?
