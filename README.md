# irw - Incremental Reading and Writing

A simple spaced repetition tool for your [incremental reading](https://en.wikipedia.org/wiki/Incremental_reading) and [incremental writing](https://supermemo.guru/wiki/Incremental_writing). Track files, review them on schedule, and let the algorithm handle timing.

## Install

```bash
go install github.com/psto/irw@latest
```

## Quick Start

```bash
irw track ~/papers/attention.pdf         # Add to reading queue
irw track ~/drafts/chapter.md -q writing # Add to writing queue
irw review                               # Start reviewing due items
irw schedule                             # See what's due
```

Press **Enter** after reviewing to schedule the next interval. Use **f** to finish an item forever, **s** to skip 1 hour, **z** to postpone 1 week.

## Configuration

Config file lives at `~/.config/irw/config.json` and is created automatically with defaults.

| Option | Default | Description |
|--------|---------|-------------|
| `db_path` | `~/.local/share/irw-tool/irw.db` | SQLite database path |
| `launcher` | `xdg-open` (Linux) | Command to open files |
| `default_queue` | `reading` | Default queue for track/review/stats commands |
| `zk_tags` | reading/writing | Map of queue names to zk tags for import |

### Custom Queues

Add custom queues by adding keys to `zk_tags`:

```json
{
  "default_queue": "reading",
  "zk_tags": {
    "reading": ["status/reading"],
    "writing": ["status/writing"],
    "research": ["status/research"]
  }
}
```

- Queue names are defined by `zk_tags` keys
- `default_queue` determines which queue is used when not specified
- Invalid queue names return an error listing all configured queues

## Commands

| Command | Description |
|---------|-------------|
| `irw track <file>` | Add file to queue. Use `--queue` (`-q`) to specify queue (must be in `zk_tags`) |
| `irw untrack <file>` | Remove from queue |
| `irw complete [file]` | Mark as finished |
| `irw priority [file] [p]` | Set priority 0–100 |
| `irw review [type] [ext]` | Interactive review. `type`: queue name from `zk_tags`. `ext`: file extension filter. `--compact` for minimal UI. |
| `irw schedule` | List due items. `--raw` for CSV, `-0` for null-delimited (pipe to xargs). |
| `irw stats [type]` | Queue stats. `type`: queue name from `zk_tags` (defaults to `default_queue`) |
| `irw import` | Sync from zk notebook (tags `status/reading`, `status/writing`) and Sioyek PDF highlights |
| `irw purge` | Remove all finished items from database |

Override the database with `irw --db /path/to/db.db ...`.

## Algorithm

Each time you review an item, its interval grows by multiplying with `afactor` (default 2.0):

| Review | Interval | Next review in |
|--------|----------|----------------|
| 1st | 1 day | 2 days |
| 2nd | 2 days | 4 days |
| 3rd | 4 days | 8 days |
| 4th | 8 days | 16 days |

Formula: `new_interval = current_interval × afactor`

## Priority Score

Range **0–100**, where **lower = higher urgency** (like in [SuperMemo](https://super-memory.com/archive/help15/read.htm#Prioritization)):
- New items get a random priority between 40–60
- A 5% random boost keeps things fresh, so that some random item jumps to the front occasionally.
- Otherwise, items are sorted by **priority**, then **due date**

Set manually: `irw priority <file> <0-100>` or in the review menu.

## External Dependencies (optional)
- [zk](https://github.com/zk-org/zk): for `irw import` to read your notebook tags
- [sioyek](https://github.com/ahrm/sioyek): for `irw import` to pick up recent PDF highlights (`~/.local/share/sioyek/shared.db`)

## Tips

Assign a global shortcut to launch quickly. For example in [niri](https://github.com/YaLTeR/niri) with my [niri-focus-or-launch](https://github.com/psto/.dotfiles/blob/main/scripts/.local/bin/niri-focus-or-launch) and [irw-scratchpad](https://github.com/psto/.dotfiles/blob/main/scripts/.local/bin/irw-scratchpad) scripts:

```
binds {
  Mod+I { spawn "~/.local/bin/niri-focus-or-launch" "irw-scratchpad" "ir_monitor"; }
}

```

Add a window rule to have the `irw` menu floating in the top right corner:

```
window-rule {
  match app-id="ir_monitor"
  open-floating true
  default-floating-position x=0 y=0 relative-to="top-right"
  default-column-width { fixed 800; }
  default-window-height { fixed 100; }
}
```

## License

[MIT](./LICENSE) License © 2026 [Piotr Stojanow](https://github.com/psto)
