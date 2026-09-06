# Contributions

Contributions are very welcome - code, docs, whatever they might be! If this is
your first contribution to an Open Source project or you're a core maintainer
of multiple projects, your time and interest in contributing is most welcome.

If you're not sure where to get started, raise an issue in the project.

Ideas may be discussed, purely on their merits and issues. Our Code of Conduct
(CoC) is straightforward - it's important that contributors feel comfortable in
discussion throughout the whole process. This project has a
[Code of Conduct](CODE_OF_CONDUCT.md).

## More than 3 lines - talk to someone first

If you're planning on making a change that's more than a 3 lines, please talk to someone first. This is so that you don't waste your time on something that might not be accepted. It's also a good way to get some feedback on your idea and make sure you're on the right track.

## A PR should be one logical change

Please try to keep your pull requests small and focused. It's almost impossible to review PRs that change lots of files for lots of different reasons. If you have a big change, it's probably best to break it down into smaller, more manageable chunks, otherwise it's likely to be rejected.

## If you're not sure, ask!

Don't be afraid to ask for advice before working on a
contribution. If you're thinking about a bigger change, especially that might
affect the core working or architecture, it's almost essential to talk and ask
about what you're planning might affect things. Some of the larger future plans may not be
documented well so it's difficult to understand how your change might affect
the general direction and roadmap of this project without asking.

The preferred way to communicate is probably via Discord or GitHub issues.

## Commit authorship

Every commit in this repository must be **authored by a human** — the person who requested, reviewed, and takes responsibility for the change.

If you used an AI coding assistant to help write the change, keep yourself as the commit author and add a **`Co-authored-by`** trailer naming the agent:

```
feat: add chore completion webhook

Co-authored-by: Cursor Agent <cursoragent@cursor.com>
```

Do not commit with an agent or bot as the sole author. Review and project policy expect human attribution on the `Author` line; co-authors record assistant involvement without replacing human ownership.

## Target the `next` branch

Open pull requests against **`next`**, not `main`. The `next` branch is **commits queued for the next release**. Admins occasionally merge `next` into `main` when cutting a new release; `main` tracks what is currently released (or release-ready).

## Mechanics of submitting a pull request

When you are ready for a PR, please see the [pull request template](.github/PULL_REQUEST_TEMPLATE.md).
