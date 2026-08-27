# Product specification

The canonical product spec lives in the repository at
[`docs/SPEC.md`](../SPEC.md) (source) and is mirrored here for MkDocs.

See the repository file for the full domain model, workflows, RBAC catalog,
API outline, and milestones.

## Summary

StarApp is a family star-rewards app. Parents are the primary admins; children
sign in with their own accounts to view stars and browse rewards.

- **Stars** — internal family currency awarded by parents
- **Rewards** — catalog items (TV time, treats, privileges) with star costs
- **Ledger** — immutable record of awards, revokes, and redemptions
- **Family members** — domain profiles linked 1:1 to IAM login accounts

## Personas

| Persona | Access |
|---------|--------|
| **Parent** | Family homepage (all children), manage children and rewards, award/revoke stars, approve redemptions, Control Panel when privileged |
| **Child** | Personal homepage (own stars only), browse rewards, request redemptions; no IAM or admin surfaces |

Child accounts use credentials the parent creates at add-child time.

## Homepage

Route `/` is role-specific:

- **Parents** (`stars.view_family`) — grid of child cards: avatar, name, balance, last award
- **Children** (`stars.view_own`) — own avatar, balance, recent awards, reward shop

## Avatars

Parents upload JPEG, PNG, or WebP images (max 2 MB) for child members. Files
are stored under the config directory and served via authenticated
`GET /avatars/{member_id}`.

## MVP scope

1. Family setup and child profiles with IAM accounts
2. Parent family homepage and child personal home
3. Avatar upload for children
4. Award and revoke stars
5. Reward catalog and redemption (with optional parent approval)
6. Balance and history views
