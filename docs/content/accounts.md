# Accounts

Every person who signs in has a username and password. Family membership is
separate: a login is linked to one person in one household.

## First administrator

An empty database creates **admin** / **admin**. Change that password before
you add children. This account can open **Control Panel**, manage people,
and change Settings.

## Parents and children

| | Parent | Child |
|---|--------|--------|
| Home | Whole family | Own stars only |
| People, rewards, chores | Yes | No |
| Control Panel, Settings, users | If they are an administrator | No |
| Redeem rewards | Can award a reward to someone | Can request their own |

Add children from **Control Panel → People** with **login** enabled. That
creates both their family profile and their sign-in.

A person can exist on the star chart without a login. Use that for younger
children: a parent marks chores and redeems rewards for them.

## How access is decided

Permissions come from **groups** and **roles**, not from ticking boxes on a
single user.

1. A **role** lists permissions (for example parent vs child).
2. A **group** is given one or more roles.
3. Users belong to groups.

The install seeds:

| Group | Typical use |
|-------|-------------|
| Administrators | Full access, including Settings and user management |
| Parents | Family administration (people, rewards, chores, stars) |
| Children | Own home, rewards, and chore completions |
| Everyone | Basic ability to use the app |

You normally do not need to edit this. Adding a child through **People**
puts them in the Children group. Keep household admins in Administrators.

Advanced edits live under **Control Panel → IAM** (users, groups, roles).
Do not grant children IAM or Settings access.

## Passwords

- **Your own:** header username → **Change password**. Current password plus
  a new password of at least eight characters.
- **Someone else:** **Control Panel → IAM → Users**, then reset their
  password. Tell them the new password in person.

## API keys

From the User Control Panel, a signed-in user can create **API keys** for
scripts or [webhooks](webhooks.md) tooling. Send the key as
`Authorization: Bearer <key>`. Read-only keys cannot change data.

The browser session uses a cookie (`starapp-sid`). API keys do not use that
cookie.

## Optional single sign-on

You can extend `auth:` in `config.yaml` for JWT, trusted headers, or mTLS
via [httpauthshim](https://github.com/jamesread/httpauthshim). Strip inbound
identity headers at your reverse proxy unless you trust the source.

Most families only need username and password.
