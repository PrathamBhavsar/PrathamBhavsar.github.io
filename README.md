# PrathamBhavsar.github.io

Static portfolio. No build step — GitHub Pages serves the files as they are.

## Adding or editing a project

1. Add an entry to `projects.json`.
2. Write the case study at `content/<slug>.md` (`## heading`, paragraphs, `- bullets`, `**bold**`, `` `code` `` — that's the whole supported subset).
3. Record a demo, then: `./tools/make-gif.sh <slug> ~/path/to/recording.mov [start] [duration]`
   It writes `media/<slug>/demo.gif` + `poster.jpg` and prints the `media` block to paste back into `projects.json`.

`projects.json` is the only source of truth for cards, links and metadata. No code change needed.

## Link policy

Three tiers, rendered from `links[].kind`:

| kind | badge | use for |
|---|---|---|
| `live` | green globe | a URL that is up right now |
| `source` | repo icon | a public GitHub repo |
| *(empty array)* | dashed lock | private or client work — the GIF is the proof |

`.github/workflows/linkcheck.yml` re-checks every URL in `projects.json` weekly and on push. A dead link fails the build.

## Local preview

```bash
python3 -m http.server 8899   # fetch() needs http://, file:// won't work
```
