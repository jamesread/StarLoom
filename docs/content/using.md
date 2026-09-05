# Using StarLoom

Daily use is short: mark chores, award extra stars, and approve rewards.

## Parent home

Signed in as a parent, **Home** shows every person in the family with their
star balance and last award. A banner links to pending redemption requests
when children have asked for rewards.

Open a person card to see their ledger, award or revoke stars, and edit
their profile.

## Child home

Signed in as a child, **Home** shows only that child’s stars:

- Current balance
- Recent awards
- Rewards they can redeem (or how many more stars they need)
- Star charts that include their chores

Rewards that need approval show as **Pending approval**. Rewards that are
not available right now (for example a weekday-only treat) are marked
unavailable.

## Star charts

Open **Star Charts** and pick a chart. Each row is a chore; each column is a
day of this week.

- A scheduled day you can complete shows as a button
- A completed day shows the stars earned
- Use the week arrows to look at previous or next weeks
- Parents can add another chore from the chart toolbar

Completing a chore writes an award to that child’s ledger.

## Awarding stars by hand

From a person’s page, enter an amount and an optional note (“Helped with
dinner”). The default amount comes from
[Settings](configuration.md) (`Default award stars`).

Revoke is for mistakes. It cannot take the balance below zero.

## Redeeming rewards

**Children** tap a reward they can afford on their home. If the reward
needs approval, stars are not spent until a parent approves it.

**Parents** can award a reward from **Rewards** (the Award action on a row)
and choose who receives it.

### Approving requests

Open **Rewards**. The **Redemption requests** table lists pending items.
Approve to deduct stars and grant the privilege; Reject leaves the balance
unchanged.

## Managing rewards

On **Rewards**, parents can add, edit, or retire catalog items. Inactive
rewards stay in history but no longer appear in the shop.

Optional **availability expression** limits when a reward can be redeemed
(for example not on Mondays). Leave it blank for always available.

## People and avatars

**Control Panel → People** lists everyone. Open **Manage** to:

- Change the display name or star colour
- Upload a JPEG, PNG, or WebP avatar (max 2 MB)
- Review the star ledger

Only parents manage avatars. Children see their own picture on their home
page.

## Passive displays

A wall tablet or a Home Assistant iframe cannot type a password. Open your
own page from the header username and copy the value under **Bearer
token**. Append `?token=<value>` to any StarLoom URL to sign that request
in as you, read-only, so a screen with no keyboard still shows a star
chart. The token is a credential, so treat the whole URL as one.
**Regenerate** issues a new token and kills the old one, which breaks every
display still holding it.
